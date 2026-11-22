import { defineMain } from "@storybook/react-vite/node"

const config = defineMain({
  stories: ["../src/**/*.mdx", "../src/**/*.stories.@(ts|tsx)"],
  addons: ["@storybook/addon-docs", "@storybook/addon-a11y", "@storybook/addon-themes"],
  framework: "@storybook/react-vite",
})

export default config
