import { defineConfig } from "vite";
import { dirname, resolve } from "path";
import { fileURLToPath } from "url";
import { readFileSync, writeFileSync, mkdirSync, copyFileSync } from "fs";

const __dirname = dirname(fileURLToPath(import.meta.url));

function extensionPlugin() {
  return {
    name: "browser-extension",
    writeBundle() {
      const pkg = JSON.parse(
        readFileSync(resolve(__dirname, "package.json"), "utf-8"),
      );
      const manifest = JSON.parse(
        readFileSync(resolve(__dirname, "src/manifest.json"), "utf-8"),
      );
      const distDir = resolve(__dirname, "dist");

      // Chrome manifest
      const chrome = JSON.parse(JSON.stringify(manifest));
      chrome.version = pkg.version;
      chrome.background.service_worker = "background.js";
      delete chrome.chrome_settings_overrides;
      writeFileSync(resolve(distDir, "manifest.json"), JSON.stringify(chrome));

      // Firefox manifest
      const ff = JSON.parse(JSON.stringify(manifest));
      ff.version = pkg.version;
      ff.background.scripts = ["background.js"];
      ff.content_security_policy = { extension_pages: "script-src 'self'" };
      const geckoSettings = {
        id: "{f0bda7ce-0cda-42dc-9ea8-126b20fed280}",
        strict_min_version: "110.0",
        data_collection_permissions: {
          required: ["browsingActivity", "websiteContent"],
        },
      };
      ff.browser_specific_settings = {
        gecko: geckoSettings,
        gecko_android: geckoSettings,
      };
      writeFileSync(resolve(distDir, "manifest_ff.json"), JSON.stringify(ff));

      // Copy static files
      copyFileSync(
        resolve(__dirname, "src/popup/popup.html"),
        resolve(distDir, "popup.html"),
      );
      copyFileSync(
        resolve(__dirname, "src/options/options.html"),
        resolve(distDir, "options.html"),
      );
      copyFileSync(
        resolve(__dirname, "src/style.css"),
        resolve(distDir, "style.css"),
      );
      mkdirSync(resolve(distDir, "assets/icons"), { recursive: true });
      copyFileSync(
        resolve(__dirname, "assets/icon128.png"),
        resolve(distDir, "assets/icons/icon128.png"),
      );
      copyFileSync(
        resolve(__dirname, "assets/logo.png"),
        resolve(distDir, "assets/logo.png"),
      );
    },
  };
}

export default defineConfig(({ mode }) => ({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: true,
    minify: mode === "production",
    rollupOptions: {
      input: {
        background: resolve(__dirname, "src/background/background.ts"),
        content: resolve(__dirname, "src/content/content.ts"),
        popup: resolve(__dirname, "src/popup/popup.ts"),
      },
      output: {
        entryFileNames: "[name].js",
        chunkFileNames: "[name].js",
        assetFileNames: "[name].[ext]",
      },
    },
  },
  plugins: [extensionPlugin()],
}));
