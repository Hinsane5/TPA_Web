import pluginVue from "eslint-plugin-vue";
import vueTsEslintConfig from "@vue/eslint-config-typescript";
import skipFormatting from "@vue/eslint-config-prettier/skip-formatting";

export default [
  {
    name: "app/files-to-lint",
    files: ["**/*.{ts,mts,tsx,vue}"],
  },

  {
    name: "app/files-to-ignore",
    ignores: [
      "**/dist/**",
      "**/dist-ssr/**",
      "**/coverage/**",
      "**/node_modules/**",
    ],
  },

  ...pluginVue.configs["flat/recommended"],

  ...vueTsEslintConfig(),

  {
    rules: {
      "vue/multi-word-component-names": "error", 
      "vue/no-v-html": "warn", 
      "vue/component-api-style": ["error", ["script-setup", "composition"]], 
      "vue/block-order": ["error", { order: ["script", "template", "style"] }], 

      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_" },
      ], 
      "@typescript-eslint/explicit-function-return-type": "off", 
      "@typescript-eslint/no-explicit-any": "warn", 

      "no-console": process.env.NODE_ENV === "production" ? "warn" : "off",
      "no-debugger": process.env.NODE_ENV === "production" ? "warn" : "off",
    },
  },

  skipFormatting,
];
