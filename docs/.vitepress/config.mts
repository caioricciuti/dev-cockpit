import { defineConfig } from "vitepress";

export default defineConfig({
  title: "Dev Cockpit",
  description: "Modern TUI for macOS & Linux developers",

  // Ignore localhost links in examples
  ignoreDeadLinks: [/^http:\/\/localhost/],

  head: [
    ["link", { rel: "icon", type: "image/png", href: "/logo.png" }],
    ["link", { rel: "shortcut icon", type: "image/png", href: "/logo.png" }],
    ["meta", { name: "theme-color", content: "#3f5fffff" }],
    [
      "script",
      {
        defer: "",
        src: "https://a.caioricciuti.com/b.js",
        "data-website-id": "4322d182-2ae0-40c4-ba2f-0587de08d1ee",
      },
    ],
  ],

  themeConfig: {
    logo: "/logo.png",

    nav: [
      { text: "Home", link: "/" },
      { text: "Getting Started", link: "/getting-started" },
      { text: "Donate", link: "https://buymeacoffee.com/caioricciuti" },
    ],

    sidebar: [
      {
        text: "Getting Started",
        items: [
          { text: "Quick Start", link: "/getting-started" },
          { text: "Troubleshooting", link: "/troubleshooting" },
        ],
      },
      {
        text: "About",
        items: [
          {
            text: "Changelog",
            link: "https://github.com/caioricciuti/dev-cockpit/releases",
          },
          { text: "Contributing", link: "/contributing" },
          { text: "Acknowledgments", link: "/acknowledgments" },
          { text: "License", link: "/license" },
        ],
      },
    ],

    socialLinks: [
      { icon: "github", link: "https://github.com/caioricciuti/dev-cockpit" },
    ],

    footer: {
      message: "Released under GPL 3.0.",
      copyright: "Copyright © 2025 Caio Ricciuti and Ibero Data",
    },

    search: {
      provider: "local",
    },
  },
});
