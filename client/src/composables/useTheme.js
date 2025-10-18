import { ref } from "vue";

const THEME_KEY = "app-theme";
const theme = ref("dark");
let initialized = false;

const applyTheme = (value) => {
  if (typeof document === "undefined") return;
  document.body.classList.toggle("light-mode", value === "light");
};

const initTheme = () => {
  if (initialized) return;
  if (typeof window !== "undefined") {
    const stored = window.localStorage.getItem(THEME_KEY);
    theme.value = stored === "light" ? "light" : "dark";
  }
  applyTheme(theme.value);
  initialized = true;
};

export function useTheme() {
  if (!initialized) {
    initTheme();
  }

  const setTheme = (value) => {
    const next = value === "light" ? "light" : "dark";
    theme.value = next;
    if (typeof window !== "undefined") {
      window.localStorage.setItem(THEME_KEY, next);
    }
    applyTheme(next);
  };

  const toggleTheme = () => {
    setTheme(theme.value === "dark" ? "light" : "dark");
  };

  return {
    theme,
    setTheme,
    toggleTheme,
  };
}
