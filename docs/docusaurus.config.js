// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

const googleTrackingId = "G-EB7MEE3TJ1";
const algoliaAppKey = "9AHLYCX3HA";
const algoliaAPIKey = "976ab1e596812cf4fbe21a3d4d1c9830";
const algoliaIndexName = "cosmos_network";

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Cosmos Hub",
  tagline: "",
  favicon: "/img/hub.svg",

  // Set the production url of your site here
  url: "https://hub.cosmos.network",
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: "/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "Cosmos", // Usually your GitHub org/user name.
  projectName: "Gaia", // Usually your repo name.

  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "warn",
  trailingSlash: false,

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  scripts: [
    {
      src: "https://kit.fontawesome.com/401fb1e734.js",
      crossorigin: "anonymous",
    },
  ],

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          versions: {
            current: {
              path: "main",
              label: "Latest",
              banner: "none",
            },
          },
        },
        sitemap: {
          changefreq: "weekly",
          priority: 0.5,
          ignorePatterns: ["/tags/**"],
          filename: "sitemap.xml",
        },
        blog: false,
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        gtag: {
          trackingID: googleTrackingId,
          anonymizeIP: true,
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: "img/banner.jpg",
      docs: {
        sidebar: {
          autoCollapseCategories: true,
          hideable: true,
        },
      },
      navbar: {
        title: "Cosmos Hub",
        hideOnScroll: false,
        logo: {
          alt: "Cosmos Hub Logo",
          src: "img/hub.svg",
          href: "/",
          target: "_self",
        },
        items: [
          {
            type: "dropdown",
            label: "Community",
            position: "right",
            items: [
              {
                href: "https://github.com/cosmos/gaia",
                html: '<i class="fa-fw fa-brands fa-github"></i> Github',
              },
              {
                href: "https://twitter.com/cosmoshub",
                html: '<i class="fa-fw fa-brands fa-twitter"></i> Twitter',
              },
              {
                href: "https://discord.gg/interchain",
                html: '<i class="fa-fw fa-brands fa-discord"></i> Discord',
              },
              {
                href: "https://forum.cosmos.network/",
                html: '<i class="fa-fw fa-regular fa-comments"></i> Forum',
              },
              {
                href: "https://reddit.com/r/cosmosnetwork",
                html: '<i class="fa-fw fa-brands fa-reddit"></i> Reddit',
              },
              {
                href: "https://www.youtube.com/c/CosmosProject",
                html: '<i class="fa-fw fa-brands fa-youtube"></i> YouTube',
              },
            ],
          },
          {
            type: "docsVersionDropdown",
            position: "left",
            // versions not yet migrated to docusaurus
            dropdownItemsAfter: [
              {
                href: "https://github.com/cosmos/gaia/tree/legacy-docs",
                label: "Archive",
                target: "_blank",
              },
            ],
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            items: [
              {
                html: `<a href="https://cosmos.network"><img src="/img/logo-bw-inverse.svg" alt="Cosmos Logo"></a>`,
              },
            ],
          },
          {
            title: "Documentation",
            items: [
              {
                label: "Cosmos SDK",
                href: "https://docs.cosmos.network/",
              },
              {
                label: "CometBFT",
                href: "https://docs.cometbft.com/",
              },
              {
                label: "IBC-Go",
                href: "https://ibc.cosmos.network/",
              },
              {
                label: "Interchain Security",
                href: "https://cosmos.github.io/interchain-security/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Forum",
                href: "https://forum.cosmos.network/",
              },
              {
                label: "Discord",
                href: "https://discord.gg/interchain",
              },
              {
                label: "Reddit",
                href: "https://reddit.com/r/cosmosnetwork",
              },
              {
                label: "Blog",
                href: "https://blog.cosmos.network/",
              },
            ],
          },
          {
            title: "Social",
            items: [
              {
                label: "Twitter",
                href: "https://twitter.com/cosmoshub",
              },
              {
                label: "Youtube",
                href: "https://www.youtube.com/c/CosmosProject",
              },
              {
                label: "Telegram",
                href: "https://t.me/cosmosproject",
              },
            ],
          },
        ],
        copyright: `This website is maintained by Interchain Foundation & Informal Systems. The contents and opinions of this website are those of Interchain Foundation & Informal Systems.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ["protobuf", "go-module"], // https://prismjs.com/#supported-languages
      },
      // algolia: {
      //   appId: algoliaAppKey,
      //   apiKey: algoliaAPIKey,
      //   indexName: algoliaIndexName,
      //   contextualSearch: false,
      // },
    }),
  themes: ["@you54f/theme-github-codeblock"],
  plugins: [
    require.resolve("docusaurus-lunr-search"),
    async function myPlugin(context, options) {
      return {
        name: "docusaurus-tailwindcss",
        configurePostCss(postcssOptions) {
          postcssOptions.plugins.push(require("postcss-import"));
          postcssOptions.plugins.push(require("tailwindcss/nesting"));
          postcssOptions.plugins.push(require("tailwindcss"));
          postcssOptions.plugins.push(require("autoprefixer"));
          return postcssOptions;
        },
      };
    },
  ],
};

module.exports = config;
