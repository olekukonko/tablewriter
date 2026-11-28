import { defineConfig } from "sourcey";

export default defineConfig({
  name: "TableWriter",
  // GitHub Pages serves project sites under /<repo>, so all links carry the
  // "/tablewriter" prefix. If the docs are ever hosted at a domain root,
  // remove this line.
  baseUrl: "/tablewriter",
  theme: {
    preset: "default",
    colors: {
      primary: "#0F766E",
    },
    fonts: {
      sans: "Inter",
      mono: "monospace",
    },
  },
  repo: "https://github.com/olekukonko/tablewriter",
  editBranch: "master",
  editBasePath: "docs",
  navigation: {
    tabs: [
      {
        tab: "Documentation",
        slug: "",
        groups: [
          { group: "Getting Started", pages: ["introduction", "quickstart"] },
          {
            group: "Guides",
            pages: [
              "configuration",
              "rendering",
              "data-input",
              "renderers",
              "width-truncation",
              "tw-package",
            ],
          },
          { group: "Reference", pages: ["cli", "migration"] },
        ],
      },
      {
        tab: "Go API",
        godoc: {
          module: "..",
          snapshot: "godoc.json",
          mode: "snapshot",
          packages: [
            "./",
            "./tw",
            "./renderer",
            "./pkg/twwidth",
            "./pkg/twcache",
            "./pkg/twwarp",
          ],
          includeTests: true,
        },
      },
    ],
  },
  navbar: {
    links: [
      { type: "github", href: "https://github.com/olekukonko/tablewriter" },
    ],
  },
  footer: {
    links: [
      { type: "github", href: "https://github.com/olekukonko/tablewriter" },
    ],
  },
  search: {
    featured: ["introduction", "quickstart"],
  },
});
