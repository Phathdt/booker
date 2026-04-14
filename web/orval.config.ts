import { defineConfig } from "orval";

export default defineConfig({
  booker: {
    input: {
      target: "../docs/openapi.yaml",
    },
    output: {
      mode: "tags-split",
      target: "src/core/api/generated",
      schemas: "src/core/api/generated/models",
      client: "react-query",
      httpClient: "axios",
      override: {
        mutator: {
          path: "src/core/api/axios-instance.ts",
          name: "axiosInstance",
        },
      },
    },
  },
  "booker-zod": {
    input: {
      target: "../docs/openapi.yaml",
    },
    output: {
      mode: "tags-split",
      target: "src/core/api/generated",
      client: "zod",
      fileExtension: ".zod",
    },
  },
});
