// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

// const lastVersion = "v11.0.0";
const googleTrackingId = 'G-EB7MEE3TJ1';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Cosmos Hub',
  tagline: '',
  favicon: 'img/hub.svg',

  // Set the production url of your site here
  url: 'https://hub.cosmos.network',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'Cosmos', // Usually your GitHub org/user name.
  projectName: 'Gaia', // Usually your repo name.

  onBrokenLinks: 'warn',
  onBrokenMarkdownLinks: 'warn',
  trailingSlash: false,

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js'),
          // lastVersion: lastVersion,
          versions: {
            current: {
              path: 'main',
              banner: 'unreleased',
            },
          },
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
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
      image: 'img/banner.jpg',
      docs: {
        sidebar: {
          autoCollapseCategories: true,
          hideable: true,
        },
      },
      navbar: {
        title: 'Cosmos Hub',
        hideOnScroll: false,
        logo: {
          alt: 'Cosmos Hub Logo',
          src: 'img/hub.svg',
          href: 'https://hub.cosmos.network',
          target: '_self',
        },
        items: [
          {
            href: 'https://github.com/cosmos/gaia',
            ico: 'img/hub.svg',
            position: 'right',
          },
          {
            href: 'https://reddit.com/r/cosmosnetwork',
            html: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="24" height="24" viewBox="0 0 256 256" xml:space="preserve">
            <g transform="matrix(1 0 0 1 128 128)" id="rE8GErn24cgvR0VS68FHy"  >
            <g style=""   >
                <g transform="matrix(2.81 0 0 2.81 -0.1434065934 -0.1434065934)" id="qWSEAI8QrJJes3kTSpPP4"  >
            <path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(255,69,0); fill-rule: nonzero; opacity: 1;"  transform=" translate(-45, -45)" d="M 0 45 C 0 20.14719 20.14719 0 45 0 C 69.85281 0 90 20.14719 90 45 C 90 69.85281 69.85281 90 45 90 C 20.14719 90 0 69.85281 0 45 z" stroke-linecap="round" />
            </g>
                <g transform="matrix(2.81 0 0 2.81 -0.0223530542 -1.274454786)" id="qbmI-Lr3R-hnjmV6cssqs"  >
            <path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 10; fill: rgb(255,255,255); fill-rule: nonzero; opacity: 1;"  transform=" translate(-45.0430795513, -44.5974917464)" d="M 75.011 45 C 74.877 41.376 71.83399999999999 38.546 68.199 38.669 C 66.588 38.724999999999994 65.056 39.385 63.893 40.492 C 58.77 37.001999999999995 52.752 35.089 46.566 34.955 L 49.485 20.916999999999998 L 59.116 22.941999999999997 C 59.384 25.413999999999998 61.599 27.203999999999997 64.071 26.934999999999995 C 66.54299999999999 26.666999999999994 68.333 24.451999999999995 68.064 21.979999999999997 C 67.79499999999999 19.508 65.58099999999999 17.717999999999996 63.108999999999995 17.987 C 61.687999999999995 18.131999999999998 60.413 18.959999999999997 59.708999999999996 20.191 L 48.68 17.987 C 47.931 17.819 47.181 18.288999999999998 47.013 19.049999999999997 C 47.013 19.060999999999996 47.013 19.060999999999996 47.013 19.071999999999996 L 43.690999999999995 34.687 C 37.42699999999999 34.788 31.330999999999996 36.711999999999996 26.140999999999995 40.224 C 23.500999999999994 37.741 19.339999999999996 37.864 16.856999999999992 40.51499999999999 C 14.373999999999992 43.154999999999994 14.496999999999993 47.315999999999995 17.147999999999993 49.79899999999999 C 17.662999999999993 50.279999999999994 18.254999999999992 50.693999999999996 18.914999999999992 50.98499999999999 C 18.86999999999999 51.64499999999999 18.86999999999999 52.30499999999999 18.914999999999992 52.96499999999999 C 18.914999999999992 63.04299999999999 30.65999999999999 71.24199999999999 45.144999999999996 71.24199999999999 C 59.629999999999995 71.24199999999999 71.375 63.05399999999999 71.375 52.96499999999999 C 71.42 52.30499999999999 71.42 51.64499999999999 71.375 50.98499999999999 C 73.635 49.855 75.056 47.528 75.011 45 z M 30.011 49.508 C 30.011 47.025000000000006 32.036 45 34.519 45 C 37.001999999999995 45 39.027 47.025 39.027 49.508 C 39.027 51.99100000000001 37.002 54.016000000000005 34.519 54.016000000000005 C 32.025 53.993 30.011 51.991 30.011 49.508 z M 56.152 62.058 L 56.152 61.879 C 52.953 64.28399999999999 49.038000000000004 65.514 45.033 65.347 C 41.028 65.515 37.114000000000004 64.28399999999999 33.914 61.87899999999999 C 33.489000000000004 61.36399999999999 33.567 60.59299999999999 34.082 60.16799999999999 C 34.529 59.79899999999999 35.167 59.79899999999999 35.626 60.16799999999999 C 38.333 62.14799999999999 41.632999999999996 63.154999999999994 44.988 62.99799999999999 C 48.344 63.17699999999999 51.655 62.21499999999999 54.394999999999996 60.25799999999999 C 54.88699999999999 59.77699999999999 55.69199999999999 59.78799999999999 56.17399999999999 60.27999999999999 C 56.655 60.772 56.644 61.577 56.152 62.058 z M 55.537 54.34 C 55.458999999999996 54.34 55.391999999999996 54.34 55.313 54.34 L 55.347 54.172000000000004 C 52.864000000000004 54.172000000000004 50.839 52.147000000000006 50.839 49.664 C 50.839 47.181 52.864 45.156 55.347 45.156 C 57.830000000000005 45.156 59.855000000000004 47.181 59.855000000000004 49.664 C 59.955 52.148 58.02 54.239 55.537 54.34 z" stroke-linecap="round" />
            </g>
            </g>
            </g>
            </svg>`,
            position: 'right',
          },
          {
            href: 'https://www.youtube.com/c/CosmosProject',
            html: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="34.69091" height="24" viewBox="-0.000004166666684568554 0 318.00000833333337 220" xml:space="preserve">
            <g transform="matrix(1 0 0 1 159 110)" id="OszDE9k35rD2P5K8ysEi-"  >
              <image style="stroke: none; stroke-width: 0; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(0,0,0); fill-rule: nonzero; opacity: 1;" vector-effect="non-scaling-stroke"  xlink:href="data:image/png;base64,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" x="-159" y="-110" width="318" height="220"></image>
            </g>
            </svg>`,
            position: 'right',
          },
          {
            href: 'https://discord.gg/cosmosnetwork',
            html: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="24" height="24" viewBox="0 0 800 800" xml:space="preserve">
            <g transform="matrix(3.125 0 0 3.125 400 400)" id="CDqljScowryMcjUf1D4XT"  >
            <path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(88,101,242); fill-rule: nonzero; opacity: 1;"  transform=" translate(-127.9999996983, -99.084465)" d="M 216.856339 16.5966031 C 200.285002 8.84328665 182.566144 3.2084988 164.041564 0 C 161.766523 4.11318106 159.108624 9.64549908 157.276099 14.0464379 C 137.583995 11.0849896 118.072967 11.0849896 98.7430163 14.0464379 C 96.9108417 9.64549908 94.1925838 4.11318106 91.8971895 0 C 73.3526068 3.2084988 55.6133949 8.86399117 39.0420583 16.6376612 C 5.61752293 67.146514 -3.4433191 116.400813 1.08711069 164.955721 C 23.2560196 181.510915 44.7403634 191.567697 65.8621325 198.148576 C 71.0772151 190.971126 75.7283628 183.341335 79.7352139 175.300261 C 72.104019 172.400575 64.7949724 168.822202 57.8887866 164.667963 C 59.7209612 163.310589 61.5131304 161.891452 63.2445898 160.431257 C 105.36741 180.133187 151.134928 180.133187 192.754523 160.431257 C 194.506336 161.891452 196.298154 163.310589 198.110326 164.667963 C 191.183787 168.842556 183.854737 172.420929 176.223542 175.320965 C 180.230393 183.341335 184.861538 190.991831 190.096624 198.16893 C 211.238746 191.588051 232.743023 181.531619 254.911949 164.955721 C 260.227747 108.668201 245.831087 59.8662432 216.856339 16.5966031 Z M 85.4738752 135.09489 C 72.8290281 135.09489 62.4592217 123.290155 62.4592217 108.914901 C 62.4592217 94.5396472 72.607595 82.7145587 85.4738752 82.7145587 C 98.3405064 82.7145587 108.709962 94.5189427 108.488529 108.914901 C 108.508531 123.290155 98.3405064 135.09489 85.4738752 135.09489 Z M 170.525237 135.09489 C 157.88039 135.09489 147.510584 123.290155 147.510584 108.914901 C 147.510584 94.5396472 157.658606 82.7145587 170.525237 82.7145587 C 183.391518 82.7145587 193.761324 94.5189427 193.539891 108.914901 C 193.539891 123.290155 183.391518 135.09489 170.525237 135.09489 Z" stroke-linecap="round" />
            </g>
            </svg>`,
            position: 'right',
          },
          {
            href: 'https://forum.cosmos.network/',
            html: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="24" height="24" viewBox="0 0 260 260" xml:space="preserve">
            <g transform="matrix(1 0 0 1 130 130)" id="FF-RWAdDVVQKW2xixfcQq" clip-path="url(#CLIPPATH_4)"  >
            <clipPath id="CLIPPATH_4" >
              <rect id="clip0" x="-130" y="-130" rx="0" ry="0" width="260" height="260" />
            </clipPath>
            <path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(0,0,0); fill-rule: evenodd; opacity: 1;"  transform=" translate(-130, -129.8204885559)" d="M 161.69 76.4992 L 77.7047 160.372 C 76.6218 158.273 75.6441 156.087 74.7683 153.819 C 71.8353 146.198 70.3688 138.113 70.3688 129.573 C 70.3688 121.027 71.8353 112.945 74.7683 105.324 C 77.7047 97.7003 81.7716 91.0864 86.9723 85.4725 C 92.1729 79.8621 98.4226 75.4245 105.718 72.1566 C 113.016 68.892 121.024 67.258 129.745 67.258 C 138.47 67.258 146.522 68.9326 153.899 72.2854 C 156.642 73.5295 159.239 74.933 161.69 76.4992 Z M 154.279 187.488 C 146.98 190.756 138.972 192.387 130.251 192.387 C 121.527 192.387 113.475 190.712 106.094 187.363 C 103.355 186.115 100.758 184.712 98.3104 183.142 L 182.292 99.2765 C 183.375 101.372 184.353 103.555 185.228 105.826 C 188.161 113.447 189.628 121.532 189.628 130.075 C 189.628 138.618 188.161 146.699 185.228 154.324 C 182.292 161.944 178.225 168.558 173.024 174.169 C 167.824 179.779 161.577 184.22 154.279 187.488 Z M 166.985 213.493 C 178.225 208.635 187.954 202.062 196.173 193.767 C 204.391 185.478 210.763 175.803 215.292 164.748 C 219.824 153.693 222.088 141.967 222.088 129.573 C 222.088 117.176 219.824 105.45 215.292 94.3916 C 212.607 87.8353 209.3 81.7808 205.376 76.2245 L 260 21.6758 L 238.294 -0.0000228882 L 183.504 54.7149 C 178.585 51.3655 173.245 48.4705 167.488 46.0262 C 156.248 41.2531 143.837 38.8631 130.251 38.8631 C 116.662 38.8631 104.251 41.2938 93.0114 46.1517 C 81.7715 51.0096 72.0423 57.5828 63.8237 65.8748 C 55.6052 74.1668 49.2299 83.8419 44.7014 94.8967 C 40.1729 105.952 37.9086 117.678 37.9086 130.075 C 37.9086 142.469 40.1729 154.192 44.7014 165.25 C 47.39 171.809 50.6964 177.864 54.6207 183.424 L 0 237.966 L 21.7057 259.641 L 76.4928 204.93 C 81.4117 208.279 86.7515 211.174 92.509 213.619 C 103.749 218.392 116.16 220.778 129.745 220.778 C 143.334 220.778 155.745 218.348 166.985 213.493 L 166.985 213.493 Z" stroke-linecap="round" />
            </g>
            </svg>`,
            position: 'right',
          },
          {
            href: 'https://github.com/cosmos/gaia',
            html: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="github-icon">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M12 0.300049C5.4 0.300049 0 5.70005 0 12.3001C0 17.6001 3.4 22.1001 8.2 23.7001C8.8 23.8001 9 23.4001 9 23.1001C9 22.8001 9 22.1001 9 21.1001C5.7 21.8001 5 19.5001 5 19.5001C4.5 18.1001 3.7 17.7001 3.7 17.7001C2.5 17.0001 3.7 17.0001 3.7 17.0001C4.9 17.1001 5.5 18.2001 5.5 18.2001C6.6 20.0001 8.3 19.5001 9 19.2001C9.1 18.4001 9.4 17.9001 9.8 17.6001C7.1 17.3001 4.3 16.3001 4.3 11.7001C4.3 10.4001 4.8 9.30005 5.5 8.50005C5.5 8.10005 5 6.90005 5.7 5.30005C5.7 5.30005 6.7 5.00005 9 6.50005C10 6.20005 11 6.10005 12 6.10005C13 6.10005 14 6.20005 15 6.50005C17.3 4.90005 18.3 5.30005 18.3 5.30005C19 7.00005 18.5 8.20005 18.4 8.50005C19.2 9.30005 19.6 10.4001 19.6 11.7001C19.6 16.3001 16.8 17.3001 14.1 17.6001C14.5 18.0001 14.9 18.7001 14.9 19.8001C14.9 21.4001 14.9 22.7001 14.9 23.1001C14.9 23.4001 15.1 23.8001 15.7 23.7001C20.5 22.1001 23.9 17.6001 23.9 12.3001C24 5.70005 18.6 0.300049 12 0.300049Z" fill="currentColor"/>
            </svg>
            `,
            position: 'right',
          },
          {
            type: 'docsVersionDropdown',
            position: 'left',
            dropdownActiveClassDisabled: true,
            // versions not yet migrated to docusaurus
            dropdownItemsAfter: [
              {
                href: 'https://hub.cosmos.network/v11/',
                label: 'v11',
                target: '_self',
              },
              {
                href: 'https://hub.cosmos.network/v10/',
                label: 'v10',
                target: '_self',
              },
            ],
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            items: [
              {
                html: `<a href="https://cosmos.network"><img src="/img/logo-bw-inverse.svg" alt="Cosmos Logo"></a>`,
              },
            ],
          },
          {
            title: 'Documentation',
            items: [
              {
                label: 'Cosmos SDK',
                href: 'https://docs.cosmos.network/',
              },
              {
                label: 'CometBFT',
                href: 'https://docs.cometbft.com/',
              },
              {
                label: 'IBC Specs',
                href: 'https://github.com/cosmos/ibc',
              },
              {
                label: 'IBC Go',
                href: 'https://ibc.cosmos.network/',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Blog',
                href: 'https://blog.cosmos.network/',
              },
              {
                label: 'Forum',
                href: 'https://forum.cosmos.network/',
              },
              {
                label: 'Discord',
                href: 'https://discord.gg/cosmosnetwork',
              },
              {
                label: 'Reddit',
                href: 'https://reddit.com/r/cosmosnetwork',
              },
            ],
          },
          {
            title: 'Social',
            items: [
              {
                label: 'Discord',
                href: 'https://discord.gg/cosmosnetwork',
              },
              {
                label: 'Twitter',
                href: 'https://twitter.com/cosmoshub',
              },
              {
                label: 'Youtube',
                href: 'https://www.youtube.com/c/CosmosProject',
              },
              {
                label: 'Telegram',
                href: 'https://t.me/cosmosproject',
              },
            ],
          },
        ],
        copyright: `This website is maintained by Interchain Foundation/Informal Systems. The contents and opinions of this website are those of Interchain Foundation/Informal Systems.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['protobuf', 'go-module'], // https://prismjs.com/#supported-languages
      },
      algolia: {
        appId: 'QLS2QSP47E',
        apiKey: '4d9feeb481e3cfef8f91bbc63e090042',
        indexName: 'cosmos_network',
        contextualSearch: false,
      },
    }),
  themes: ['@you54f/theme-github-codeblock'],
  plugins: [
    async function myPlugin(context, options) {
      return {
        name: 'docusaurus-tailwindcss',
        configurePostCss(postcssOptions) {
          postcssOptions.plugins.push(require('postcss-import'));
          postcssOptions.plugins.push(require('tailwindcss/nesting'));
          postcssOptions.plugins.push(require('tailwindcss'));
          postcssOptions.plugins.push(require('autoprefixer'));
          return postcssOptions;
        },
      };
    },
    [
      '@docusaurus/plugin-client-redirects',
      {
        fromExtensions: ['html'],
        toExtensions: ['html'],
        redirects: [],
      },
    ],
  ],
};

module.exports = config;
