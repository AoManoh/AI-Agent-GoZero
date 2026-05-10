#!/usr/bin/env python3
"""将前端离线面试题库资产导入 PostgreSQL 结构化题库表。"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
import sys
from collections import Counter
from pathlib import Path
from typing import Any


DEFAULT_ASSET = Path("client/public/data/interview-question-bank.json")
DEFAULT_CONTAINER = os.environ.get("POSTGRES_CONTAINER", "gozero-ai-postgres-1")
DEFAULT_DB = os.environ.get("POSTGRES_DB", "gozero_ai_agent")
DEFAULT_USER = os.environ.get("POSTGRES_USER", "root")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="导入面试题库 JSON 到 interview_questions / interview_question_sources。",
    )
    parser.add_argument("--asset", default=str(DEFAULT_ASSET), help="题库 JSON 文件路径")
    parser.add_argument("--dry-run", action="store_true", help="只校验和统计，不写数据库")
    parser.add_argument("--preserve-status", action="store_true", help="保留 JSON 中的 ingestStatus")
    parser.add_argument("--database-url", default="", help="使用本地 psql 连接的 PostgreSQL DSN")
    parser.add_argument("--docker-container", default=DEFAULT_CONTAINER, help="未提供 DSN 时使用的 PostgreSQL 容器名")
    parser.add_argument("--db", default=DEFAULT_DB, help="容器内数据库名")
    parser.add_argument("--user", default=DEFAULT_USER, help="容器内数据库用户")
    return parser.parse_args()


def load_asset(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as file:
        data = json.load(file)
    if not isinstance(data, dict):
        raise ValueError("题库资产根节点必须是对象")
    questions = data.get("questions")
    if not isinstance(questions, list) or not questions:
        raise ValueError("题库资产 questions 必须是非空数组")
    return data


def as_text(value: Any, fallback: str = "") -> str:
    text = str(value if value is not None else "").strip()
    return text or fallback


def as_int(value: Any, fallback: int = 0) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return fallback


def clamp(value: int, lower: int, upper: int) -> int:
    return max(lower, min(upper, value))


def string_list(value: Any) -> list[str]:
    if not isinstance(value, list):
        return []
    result: list[str] = []
    seen: set[str] = set()
    for item in value:
        text = as_text(item)
        if text and text not in seen:
            result.append(text)
            seen.add(text)
    return result


def source_refs(value: Any) -> list[dict[str, str]]:
    if not isinstance(value, list):
        return []
    result: list[dict[str, str]] = []
    seen: set[tuple[str, str]] = set()
    for item in value:
        if isinstance(item, dict):
            key = as_text(item.get("key") or item.get("id") or item.get("sourceKey"))
            title = as_text(item.get("title") or item.get("name"), key)
            url = as_text(item.get("url") or item.get("href"))
            source_type = as_text(item.get("type"), "reference")
            license_note = as_text(item.get("license") or item.get("licenseNote"))
        else:
            key = as_text(item)
            title = key
            url = ""
            source_type = "reference"
            license_note = ""
        if not key:
            continue
        identity = (key, url)
        if identity in seen:
            continue
        seen.add(identity)
        result.append(
            {
                "key": key,
                "title": title,
                "url": url,
                "type": source_type,
                "license": license_note,
            }
        )
    return result


def json_literal(value: Any) -> str:
    return sql_literal(json.dumps(value, ensure_ascii=False, separators=(",", ":"))) + "::jsonb"


def sql_literal(value: Any) -> str:
    if value is None:
        return "NULL"
    text = str(value)
    return "'" + text.replace("'", "''") + "'"


def content_hash(question: dict[str, Any]) -> str:
    payload = {
        "title": as_text(question.get("title")),
        "prompt": as_text(question.get("prompt")),
        "expectedSignals": string_list(question.get("expectedSignals")),
        "followUps": string_list(question.get("followUps")),
        "evaluationDimensions": string_list(question.get("evaluationDimensions")),
    }
    raw = json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def normalize_question(raw: dict[str, Any], index: int, preserve_status: bool) -> dict[str, Any]:
    key = as_text(raw.get("key"))
    title = as_text(raw.get("title"))
    prompt = as_text(raw.get("prompt"))
    if not key or not title or not prompt:
        raise ValueError(f"第 {index + 1} 道题缺少 key/title/prompt")

    difficulty = clamp(as_int(raw.get("difficultyLevel"), 3), 1, 5)
    status = as_text(raw.get("ingestStatus"), "ready") if preserve_status else "ready"
    if status not in {"ready", "draft", "archived"}:
        status = "ready"

    return {
        "key": key,
        "direction_key": as_text(raw.get("directionKey"), "go_backend"),
        "focus_key": as_text(raw.get("focusKey"), "engineering"),
        "focus_label": as_text(raw.get("focusLabel")),
        "difficulty_level": difficulty,
        "difficulty_label": as_text(raw.get("difficultyLabel")),
        "title": title,
        "prompt": prompt,
        "expected_signals": string_list(raw.get("expectedSignals")),
        "follow_ups": string_list(raw.get("followUps")),
        "evaluation_dimensions": string_list(raw.get("evaluationDimensions")),
        "tags": string_list(raw.get("qualityTags")),
        "source_refs_raw": string_list(raw.get("sourceRefs")),
        "source_refs": source_refs(raw.get("sourceRefs")),
        "batch_key": as_text(raw.get("batch")),
        "batch_label": as_text(raw.get("batchLabel")),
        "sequence": as_int(raw.get("sequence"), index + 1),
        "batch_sequence": as_int(raw.get("batchSequence"), 0),
        "status": status,
        "quality_score": min(
            100,
            50
            + len(string_list(raw.get("expectedSignals"))) * 6
            + len(string_list(raw.get("followUps"))) * 4
            + len(string_list(raw.get("evaluationDimensions"))) * 3,
        ),
        "content_hash": content_hash(raw),
    }


def build_question_sql(question: dict[str, Any]) -> str:
    fields = [
        "question_key",
        "direction_key",
        "focus_key",
        "focus_label",
        "difficulty_level",
        "difficulty_label",
        "title",
        "prompt",
        "expected_signals",
        "follow_ups",
        "evaluation_dimensions",
        "tags",
        "source_refs",
        "batch_key",
        "batch_label",
        "sequence",
        "batch_sequence",
        "status",
        "quality_score",
        "content_hash",
    ]
    values = [
        sql_literal(question["key"]),
        sql_literal(question["direction_key"]),
        sql_literal(question["focus_key"]),
        sql_literal(question["focus_label"]),
        str(question["difficulty_level"]),
        sql_literal(question["difficulty_label"]),
        sql_literal(question["title"]),
        sql_literal(question["prompt"]),
        json_literal(question["expected_signals"]),
        json_literal(question["follow_ups"]),
        json_literal(question["evaluation_dimensions"]),
        json_literal(question["tags"]),
        json_literal(question["source_refs_raw"]),
        sql_literal(question["batch_key"]),
        sql_literal(question["batch_label"]),
        str(question["sequence"]),
        str(question["batch_sequence"]),
        sql_literal(question["status"]),
        str(question["quality_score"]),
        sql_literal(question["content_hash"]),
    ]
    insert = f"""
INSERT INTO "public"."interview_questions" ({", ".join(fields)})
VALUES ({", ".join(values)})
ON CONFLICT (question_key) DO UPDATE SET
    direction_key = EXCLUDED.direction_key,
    focus_key = EXCLUDED.focus_key,
    focus_label = EXCLUDED.focus_label,
    difficulty_level = EXCLUDED.difficulty_level,
    difficulty_label = EXCLUDED.difficulty_label,
    title = EXCLUDED.title,
    prompt = EXCLUDED.prompt,
    expected_signals = EXCLUDED.expected_signals,
    follow_ups = EXCLUDED.follow_ups,
    evaluation_dimensions = EXCLUDED.evaluation_dimensions,
    tags = EXCLUDED.tags,
    source_refs = EXCLUDED.source_refs,
    batch_key = EXCLUDED.batch_key,
    batch_label = EXCLUDED.batch_label,
    sequence = EXCLUDED.sequence,
    batch_sequence = EXCLUDED.batch_sequence,
    status = EXCLUDED.status,
    quality_score = EXCLUDED.quality_score,
    content_hash = EXCLUDED.content_hash,
    updated_at = now()
RETURNING id
"""
    sources = question["source_refs"]
    if not sources:
        return f"""
WITH upserted AS ({insert})
DELETE FROM "public"."interview_question_sources" s
USING upserted
WHERE s.question_id = upserted.id;
"""

    source_selects = []
    for source in sources:
        source_selects.append(
            "SELECT id, "
            + ", ".join(
                [
                    sql_literal(source["key"]),
                    sql_literal(source["title"]),
                    sql_literal(source["url"]),
                    sql_literal(source["type"]),
                    sql_literal(source["license"]),
                    sql_literal(question["batch_key"]),
                ]
            )
            + " FROM upserted"
        )
    return f"""
WITH upserted AS ({insert}),
deleted AS (
    DELETE FROM "public"."interview_question_sources" s
    USING upserted
    WHERE s.question_id = upserted.id
)
INSERT INTO "public"."interview_question_sources"
    (question_id, source_key, source_title, source_url, source_type, license_note, batch_key)
{" UNION ALL ".join(source_selects)}
ON CONFLICT ON CONSTRAINT uq_interview_question_sources_identity DO NOTHING;
"""


def build_import_sql(questions: list[dict[str, Any]]) -> str:
    statements = ["BEGIN;", "SET LOCAL statement_timeout = '90s';"]
    statements.extend(build_question_sql(question) for question in questions)
    statements.append("COMMIT;")
    return "\n".join(statements)


def run_psql(sql: str, args: argparse.Namespace) -> None:
    if args.database_url:
        command = ["psql", args.database_url, "-v", "ON_ERROR_STOP=1"]
    else:
        command = [
            "docker",
            "exec",
            "-i",
            args.docker_container,
            "psql",
            "-U",
            args.user,
            "-d",
            args.db,
            "-v",
            "ON_ERROR_STOP=1",
        ]
    subprocess.run(command, input=sql.encode("utf-8"), check=True)


def print_summary(questions: list[dict[str, Any]]) -> None:
    by_direction = Counter(q["direction_key"] for q in questions)
    by_difficulty = Counter(q["difficulty_level"] for q in questions)
    print(f"questions={len(questions)}")
    print("directions=" + json.dumps(dict(sorted(by_direction.items())), ensure_ascii=False))
    print("difficulties=" + json.dumps(dict(sorted(by_difficulty.items())), ensure_ascii=False))
    print("first_keys=" + ", ".join(q["key"] for q in questions[:5]))


def main() -> int:
    args = parse_args()
    asset_path = Path(args.asset)
    data = load_asset(asset_path)
    questions = [
        normalize_question(raw, index, args.preserve_status)
        for index, raw in enumerate(data["questions"])
        if isinstance(raw, dict)
    ]
    if len(questions) != len(data["questions"]):
        raise ValueError("题库资产中存在非对象题目")

    print_summary(questions)
    if args.dry_run:
        return 0

    run_psql(build_import_sql(questions), args)
    print("imported=ok")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"import_failed={exc}", file=sys.stderr)
        raise SystemExit(1)
