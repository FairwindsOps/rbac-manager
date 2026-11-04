// Project-owned. The shared VitePress config loads this file.
// Sidebar groups may use the VuePress shape (title, path, children, collapsable)
// or the VitePress shape (text, link, items, collapsed).
// https://vitepress.dev/reference/default-theme-sidebar
module.exports = {
  title: "Rbac Manager Documentation",
  description: "Documentation for Fairwinds' Rbac Manager",
  themeConfig: {
    docsRepo: "FairwindsOps/rbac-manager",
    sidebar: [
      {
        title: "Rbac Manager",
        path: "/",
      },
      {
        title: "Introduction",
        path: "/introduction",
      },
      {
        title: "RBACDefinitions",
        path: "/rbacdefinitions"
      },
      {
        title: "Upgrades",
        path: "/upgrades",
      },
      {
        title: "Cloud Provider Auth",
        children: [
          {
            title: "AWS",
            path: "/aws"
          },
          {
            title: "AKS",
            path: "/aks"
          },
          {
            title: "GKE",
            path: "/gke",
          }
        ]
      },
      {
        title: "Contributing",
        children: [
          {
            title: "Guide",
            path: "contributing/guide"
          },
          {
            title: "Code of Conduct",
            path: "contributing/code-of-conduct"
          }
        ]
      }
    ]
  },
}
