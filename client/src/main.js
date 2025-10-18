import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import ElementPlus from "element-plus";
import ElementPlusX from "vue-element-plus-x";
import "element-plus/dist/index.css";
import "./style.css";

const app = createApp(App);

app.use(router);
app.use(ElementPlus);
app.use(ElementPlusX);
app.mount("#app");
