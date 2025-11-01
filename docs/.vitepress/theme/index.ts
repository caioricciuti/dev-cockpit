import DefaultTheme from "vitepress/theme";
import "./custom.css";
import ScreenshotGallery from "../components/ScreenshotGallery.vue";

export default {
  ...DefaultTheme,
  enhanceApp({ app }: { app: import("vue").App }) {
    app.component("ScreenshotGallery", ScreenshotGallery);
  },
};
