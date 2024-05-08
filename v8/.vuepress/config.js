const p = require('../governance/current-parameters.json');

module.exports = {
  theme: "cosmos",
  title: "Cosmos Hub",
  locales: {
    "/": {
      selectText: 'Languages',
      label: "English",
      lang: "en-US"
    },
    "/es/": {
      selectText: 'Idiomas',
      label: "español",
      lang: "es"
    },
    "/ko/": {
      selectText: "언어 선택",
      label: "한국어",
      lang: "ko"
    },
    "/zh/": {
      selectText: "选择语言",
      label: "中文(简体)",
      lang: "zh-CN"
    }
  },
  base: process.env.VUEPRESS_BASE || "/",
  themeConfig: {
    docsRepo: "cosmos/gaia",
    docsBranch: "main",
    docsDir: "docs",
    editLinks: true,
    label: "hub",
    currentParameters: p,
    topbar: {
      banner: false,
    },
    sidebar: {
      nav: [
        {
          title: "Resources",
          children: [
            {
              title: "Tutorials",
              path: "https://tutorials.cosmos.network"
            },
            {
              title: "SDK API Reference",
              path: "https://godoc.org/github.com/cosmos/cosmos-sdk"
            },
            {
              title: "REST API Spec",
              path: "https://cosmos.network/rpc/"
            }
          ]
        }
      ]
    },
    gutter: {
      editLink: true,
    },
    footer: {
      question: {
        text: "Chat with Cosmos developers in <a href='https://discord.gg/cosmosnetwork' target='_blank'>Discord</a> or reach out on the <a href='https://forum.cosmos.network/c/cosmos-sdk' target='_blank'>SDK Developer Forum</a> to learn more."
      },
      logo: "/logo-bw.svg",
      textLink: {
        text: "cosmos.network",
        url: "https://cosmos.network"
      },
      services: [
        {
          service: "medium",
          url: "https://blog.cosmos.network/"
        },
        {
          service: "twitter",
          url: "https://twitter.com/cosmos"
        },
        {
          service: "linkedin",
          url: "https://www.linkedin.com/company/tendermint/"
        },
        {
          service: "github",
          url: "https://github.com/cosmos/gaia"
        },
        {
          service: "reddit",
          url: "https://reddit.com/r/cosmosnetwork"
        },
        {
          service: "telegram",
          url: "https://t.me/cosmosproject"
        },
        {
          service: "youtube",
          url: "https://www.youtube.com/c/CosmosProject"
        }
      ],
      smallprint:
        "This website is maintained by Tendermint Inc. The contents and opinions of this website are those of Tendermint Inc.",
      links: [
        {
          title: "Documentation",
          children: [
            {
              title: "Cosmos SDK",
              url: "https://docs.cosmos.network"
            },
            {
              title: "Cosmos Hub",
              url: "https://hub.cosmos.network/"
            },
            {
              title: "Tendermint Core",
              url: "https://docs.tendermint.com/"
            }
          ]
        },
        {
          title: "Community",
          children: [
            {
              title: "Cosmos blog",
              url: "https://blog.cosmos.network/"
            },
            {
              title: "Forum",
              url: "https://forum.cosmos.network/"
            }
          ]
        },
        {
          title: "Contributing",
          children: [
            {
              title: "Contributing to the docs",
              url:
                "https://github.com/cosmos/gaia/blob/main/docs/DOCS_README.md"
            },
            {
              title: "Source code on GitHub",
              url: "https://github.com/cosmos/gaia/"
            }
          ]
        }
      ]
    }
  },
  plugins: [
    [
      "@vuepress/google-analytics",
      {
        ga: "UA-51029217-2"
      }
    ],
    [
      "sitemap",
      {
        hostname: "https://hub.cosmos.network"
      }
    ],
    [ "tabs" ]
  ]
};
