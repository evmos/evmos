module.exports = {
  theme: 'cosmos',
  title: 'Evmos Documentation',
  locales: {
    '/': {
      lang: 'en-US'
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
        href:
          "https://cdnjs.cloudflare.com/ajax/libs/KaTeX/0.5.1/katex.min.css",
      },
    ],
    [
      "link",
      {
        rel: "stylesheet",
        href:
          "https://cdn.jsdelivr.net/github-markdown-css/2.2.1/github-markdown.css",
      },
    ],
  ],
  base: process.env.VUEPRESS_BASE || '/',
  plugins: [
    'vuepress-plugin-element-tabs'
  ],
  head: [
    // ['link', { rel: "apple-touch-icon", sizes: "180x180", href: "/apple-touch-icon.png" }],
    ['link', { rel: "icon", type: "image/png", sizes: "32x32", href: "/favicon32.png" }],
    ['link', { rel: "icon", type: "image/png", sizes: "16x16", href: "/favicon16.png" }],
    ['link', { rel: "manifest", href: "/site.webmanifest" }],
    ['meta', { name: "msapplication-TileColor", content: "#2e3148" }],
    ['meta', { name: "theme-color", content: "#ffffff" }],
    ['link', { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" }],
    // ['link', { rel: "apple-touch-icon-precomposed", href: "/apple-touch-icon-precomposed.png" }],
  ],
  themeConfig: {
    repo: 'tharsis/evmos',
    docsRepo: 'tharsis/evmos',
    docsBranch: 'main',
    docsDir: 'docs',
    editLinks: true,
    custom: true,
    project: {
      name: 'Evmos',
      denom: 'Evmos',
      ticker: 'EVMOS',
      binary: 'evmosd',
      testnet_denom: 'tEvmos',
      testnet_ticker: 'tEVMOS',
      rpc_url: 'https://eth.bd.evmos.org:8545',
      rpc_url_testnet: 'https://eth.bd.evmos.dev:8545',
      rpc_url_local: 'http://localhost:8545/',
      chain_id: '9001',
      testnet_chain_id: '9000',
      latest_version: 'v2.0.1',
      version_number: '2',
      testnet_version_number: '3',
      testnet_evm_explorer_url: 'https://evm.evmos.dev',
      evm_explorer_url: 'https://evm.evmos.org',
      testnet_cosmos_explorer_url: 'https://explorer.evmos.dev/',
      cosmos_explorer_url: 'https://www.mintscan.io/evmos',
    },
    logo: {
      src: '/evmos-black.svg',
    },
    algolia: {
      id: 'BH4D9OD16A',
      key: 'a5d55fe5f540cc3bd28fa2c72f2b5bd8',
      index: 'evmos'
    },
    topbar: {
      banner: false
    },
    sidebar: {
      auto: false,
      nav: [
        {
          title: 'Reference',
          children: [
            {
              title: 'Introduction',
              directory: true,
              path: '/intro'
            },
            {
              title: 'Quick Start',
              directory: true,
              path: '/quickstart'
            },
            {
              title: 'Basics',
              directory: true,
              path: '/basics'
            },
            {
              title: 'Core Concepts',
              directory: true,
              path: '/core'
            },
          ]
        },
        {
          title: 'Guides',
          children: [
            {
              title: 'Localnet',
              directory: true,
              path: '/guides/localnet'
            },
            {
              title: 'Keys and Wallets',
              directory: true,
              path: '/guides/keys-wallets'
            },
            {
              title: 'Ethereum Tooling',
              directory: true,
              path: '/guides/tools'
            },
            {
              title: 'Validators',
              directory: true,
              path: '/guides/validators'
            },
            {
              title: 'Upgrades',
              directory: true,
              path: '/guides/upgrades'
            },
            {
              title: 'Key Management System',
              directory: true,
              path: '/guides/kms'
            },
          ]
        },
        {
          title: 'APIs',
          children: [
            {
              title: 'JSON-RPC',
              directory: true,
              path: '/api/json-rpc'
            },
            {
              title: 'Protobuf Reference',
              directory: false,
              path: '/api/proto-docs'
            },
          ]
        },
        // {
        //   title: 'Clients',
        //   children: [
        //     {
        //       title: 'APIs',
        //       directory: false,
        //       path: '/clients/apis'
        //     },
        //     {
        //       title: 'Evmosjs',
        //       directory: false,
        //       path: '/clients/evmosjs'
        //     },
        //   ]
        // },
        {
          title: 'Mainnet',
          children: [
            {
              title: 'Join Mainnet',
              directory: false,
              path: '/mainnet/join'
            },
          ]
        },
        {
          title: 'Testnet',
          children: [
            {
              title: 'Join Testnet',
              directory: false,
              path: '/testnet/join'
            },
            {
              title: 'Token Faucet',
              directory: false,
              path: '/testnet/faucet'
            },
            {
              title: 'Deploy Node on Cloud',
              directory: false,
              path: '/testnet/cloud_providers'
            }
          ]
        },
        {
          title: 'Specifications',
          children: [{
            title: 'Modules',
            directory: true,
            path: '/modules'
          }]
        },
        {
          title: 'Block Explorers',
          children: [
            {
              title: 'Block Explorers',
              path: '/tools/explorers'
            },
            {
              title: 'Blockscout (EVM)',
              path: 'https://evm.evmos.org'
            },
            {
              title: 'Mintscan (Cosmos)',
              path: 'https://www.mintscan.io/evmos/'
            },
          ]
        },
        {
          title: 'Ecosystem',
          children: [
            {
              title: 'Awesome Evmos',
              path: 'https://github.com/tharsis/awesome'
            },
            {
              title: 'Evmos Space',
              path: 'https://evmos.space/'
            }
          ]
        },
        {
          title: 'Resources',
          children: [
            {
              title: 'Evmos Go API',
              path: 'https://pkg.go.dev/github.com/tharsis/evmos'
            },
            {
              title: 'Ethermint Library Go API',
              path: 'https://pkg.go.dev/github.com/tharsis/ethermint'
            },
            {
              title: 'Evmos gRPC Gateway API',
              path: 'https://api.evmos.dev/'
            },
            {
              title: 'JSON-RPC API',
              path: '/api/json-rpc/endpoints'
            }
          ]
        }
      ]
    },
    gutter: {
      title: 'Help & Support',
      chat: {
        title: 'Developer Chat',
        text: 'Chat with Evmos developers on Discord.',
        url: 'https://discord.gg/evmos',
        bg: 'linear-gradient(103.75deg, #1B1E36 0%, #22253F 100%)'
      },
      forum: {
        title: 'Evmos Developer Forum',
        text: 'Join the Evmos Developer Forum to learn more.',
        url: 'https://forum.cosmos.network/c/ethermint', // TODO: replace with commonwealth link
        bg: 'linear-gradient(221.79deg, #3D6B99 -1.08%, #336699 95.88%)',
        logo: 'ethereum-white'
      },
      github: {
        title: 'Found an Issue?',
        text: 'Help us improve this page by suggesting edits on GitHub.',
        bg: '#F8F9FC'
      }
    },
    footer: {
      logo: '/evmos-black.svg',
      textLink: {
        text: 'evmos.org',
        url: 'https://evmos.org'
      },
      services: [
        {
          service: 'github',
          url: 'https://github.com/tharsis/evmos'
        },
        {
          service: "twitter",
          url: "https://twitter.com/EvmosOrg",
        },
        {
          service: "telegram",
          url: "https://t.me/EvmosOrg",
        },
        {
          service: "linkedin",
          url: "https://www.linkedin.com/company/tharsis-finance/",
        },
        {
          service: "medium",
          url: "https://evmos.blog/",
        },
      ],
      smallprint: 'This website is maintained by Tharsis Labs Ltd.',
      links: [{
        title: 'Documentation',
        children: [{
          title: 'Cosmos SDK Docs',
          url: 'https://docs.cosmos.network/master/'
        },
        {
          title: 'Ethereum Docs',
          url: 'https://ethereum.org/developers'
        },
        {
          title: 'Tendermint Core Docs',
          url: 'https://docs.tendermint.com'
        }
        ]
      },
      {
        title: 'Community',
        children: [{
          title: 'Evmos Community',
          url: 'https://discord.gg/evmos'
        },
        ]
      },
      {
        title: 'Tharsis',
        children: [
          {
            title: 'Jobs at Tharsis',
            url: 'https://tharsis.notion.site/'
          }
        ]
      }
      ]
    },
    versions: [
      {
        "label": "main",
        "key": "main"
      },
    ],
  }
};
