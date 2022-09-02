module.exports = {
  theme: "cosmos",
  title: "Point Chain Documentation",
  locales: {
    "/": {
      lang: "en-US",
    },
  },
  markdown: {
    extendMarkdown: (md) => {
      md.use(require("markdown-it-katex"));
    },
  },
  head: [
    [
      "link",
      {
        rel: "stylesheet",
        href: "https://cdnjs.cloudflare.com/ajax/libs/KaTeX/0.5.1/katex.min.css",
      },
    ],
    [
      "link",
      {
        rel: "stylesheet",
        href: "https://cdn.jsdelivr.net/github-markdown-css/2.2.1/github-markdown.css",
      },
    ],
  ],
  base: process.env.VUEPRESS_BASE || "/",
  plugins: [
    ["vuepress-plugin-element-tabs"],
    [
      "@vuepress/google-analytics",
      {
        ga: "UA-232833231-1",
      },
    ],
  ],
  head: [
    // ['link', { rel: "apple-touch-icon", sizes: "180x180", href: "/apple-touch-icon.png" }],
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        sizes: "32x32",
        href: "/favicon32.png",
      },
    ],
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        sizes: "16x16",
        href: "/favicon16.png",
      },
    ],
    ["link", { rel: "manifest", href: "/site.webmanifest" }],
    ["meta", { name: "msapplication-TileColor", content: "#2e3148" }],
    ["meta", { name: "theme-color", content: "#ffffff" }],
    ["link", { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" }],
    // ['link', { rel: "apple-touch-icon-precomposed", href: "/apple-touch-icon-precomposed.png" }],
  ],
  themeConfig: {
    repo: "tharsis/Point Chain",
    docsRepo: "tharsis/Point Chain",
    docsBranch: "main",
    docsDir: "docs",
    editLinks: true,
    custom: true,
    project: {
      name: "Point Chain",
      denom: "POINT",
      ticker: "Point Chain",
      binary: "Point Chaind",
      testnet_denom: "tPoint Chain",
      testnet_ticker: "tPoint Chain",
      rpc_url: "https://rpc-mainnet-1.point.space",
      rpc_url_testnet: "http://xnet-uranus-1.point.space/",
      rpc_url_local: "http://localhost:8545/",
      chain_id: "10687",
      testnet_chain_id: "10731",
      latest_version: "v7.0.0",
      version_number: "1",
      testnet_version_number: "1",
      testnet_evm_explorer_url: "https://explorer.pointnetwork.io/",
      evm_explorer_url: "https://explorer.pointnetwork.io/",
      testnet_cosmos_explorer_url: "https://cosmos.pointnetwork.io/point",
      cosmos_explorer_url: "https://cosmos.pointnetwork.io/point",
    },
    logo: {
      src: "/evmos-black.svg",
    },
    algolia: {
      id: "K3VQTEW3G5",
      key: "bf836a3c934b1d4df091d5c5b69c65d7",
      index: "Point Chain",
    },
    topbar: {
      banner: false,
    },
    sidebar: {
      auto: false,
      nav: [
        {
          title: "About Point Chain",
          children: [
            {
              title: "Introduction",
              directory: true,
              path: "/about/intro",
            },
            // {
            //   title: "Evmos Ecosystem",
            //   path: "https://evmos.space/",
            // },
            // {
            //   title: "Awesome Evmos",
            //   path: "https://github.com/tharsis/awesome",
            // },
          ],
        },
        {
          title: "For Users",
          children: [
            {
              title: "Basic Concepts",
              directory: true,
              path: "/users/basics",
            },
            {
              title: "Digital Wallets",
              directory: true,
              path: "/users/wallets",
            },
            {
              title: "Account Keys",
              directory: true,
              path: "/users/keys",
            },
            {
              title: "Point Chain Governance",
              directory: true,
              path: "/users/governance",
            },
            {
              title: "Technical Concepts",
              directory: true,
              path: "/users/technical_concepts",
            },
          ],
        },
        {
          title: "For Validators",
          children: [
            {
              title: "Validators Overview",
              directory: false,
              path: "/validators/overview",
            },
            {
              title: "Installation & Quick Start",
              directory: true,
              path: "/validators/quickstart",
            },
            {
              title: "Setup & Configuration",
              directory: true,
              path: "/validators/setup",
            },
            {
              title: "Join Testnet",
              directory: false,
              path: "/validators/testnet",
            },
            {
              title: "Join Mainnet",
              directory: false,
              path: "/validators/mainnet",
            },
            {
              title: "Telemetry and Observability",
              directory: false,
              path: "/protocol/telemetry",
            },
            {
              title: "Security",
              directory: true,
              path: "/validators/security",
            },
            {
              title: "Software Upgrade Guide",
              directory: true,
              path: "/validators/upgrades",
            },
            {
              title: "Snapshots & Archive Nodes",
              directory: false,
              path: "/validators/snapshots_archives",
            },
            {
              title: "FAQ",
              directory: false,
              path: "/validators/faq",
            },
          ],
        },
        {
          title: "Block Explorers",
          children: [
            {
              title: "Block Explorers",
              path: "/developers/explorers",
            },
            {
              title: "Blockscout (EVM)",
              path: "https://evm.Point Chain.org",
            },
            {
              title: "Mintscan (Cosmos)",
              path: "https://www.mintscan.io/Point Chain/",
            },
          ],
        },
      ],
    },
    gutter: {
      title: "Help & Support",
      chat: {
        title: "Discord Channel",
        text: "Chat with Point Chain users and team on Discord.",
        url: "https://discord.com/invite/DkH6zxCXWz",
        bg: "linear-gradient(103.75deg, #1B1E36 0%, #22253F 100%)",
      },
      github: {
        title: "Found an Issue?",
        text: "Help us improve this page by suggesting edits on GitHub.",
        bg: "#F8F9FC",
      },
    },
    footer: {
      logo: "/evmos-black.svg",
      textLink: {
        text: "Point Chain",
        url: "https://pointnetwork.io/",
      },
      services: [
        {
          service: "github",
          url: "https://github.com/pointnetwork/point-chain",
        },
        {
          service: "twitter",
          url: "https://twitter.com/pointnetwork",
        },
        {
          service: "telegram",
          url: "https://t.me/pointnetworkchat",
        },
        {
          service: "linkedin",
          url: "https://www.linkedin.com/company/point-network",
        },
        {
          service: "medium",
          url: "https://medium.com/@Pointnetwork",
        },
      ],
      smallprint: "This website is maintained by Point Network.",
      links: [
        {
          title: "Ecosystem Documentation",
          children: [
            {
              title: "Cosmos SDK Docs",
              url: "https://docs.cosmos.network",
            },
            {
              title: "Ethereum Docs",
              url: "https://ethereum.org/developers",
            },
            {
              title: "Tendermint Core Docs",
              url: "https://docs.tendermint.com",
            },
          ],
        },
        {
          title: "Community",
          children: [
            {
              title: "Point Network Discord Community",
              url: "https://discord.com/invite/DkH6zxCXWz",
            }
          ],
        },
        {
          title: "Point Chain",
          children: [
            {
              title: "Jobs at Point Network",
              url: "https://pointnetwork-1639566514062.freshteam.com/jobs",
            },
          ],
        },
      ],
    },
    versions: [
      {
        label: "main",
        key: "main",
      },
    ],
  },
};
