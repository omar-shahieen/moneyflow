import { extendZodWithOpenApi } from "@anatine/zod-openapi";
import { z } from "zod";

extendZodWithOpenApi(z);
import { generateOpenApi } from "@ts-rest/open-api";

import { apiContract } from "./contracts/index.js";

type SecurityRequirementObject = {
  [key: string]: string[];
};

export type OperationMapper = NonNullable<
  Parameters<typeof generateOpenApi>[2]
>["operationMapper"];

const hasSecurity = (
  metadata: unknown
): metadata is { openApiSecurity: SecurityRequirementObject[] } => {
  return (
    !!metadata && typeof metadata === "object" && "openApiSecurity" in metadata
  );
};

const operationMapper: OperationMapper = (operation, appRoute) => ({
  ...operation,
  ...(hasSecurity(appRoute.metadata)
    ? {
        security: appRoute.metadata.openApiSecurity,
      }
    : {}),
});

export const OpenAPI = Object.assign(
  generateOpenApi(
    apiContract,
    {
      openapi: "3.0.2",
      info: {
        version: "1.0.0",
        title: "MoneyFlow REST API - Documentation",
        description: "MoneyFlow REST API - Documentation",
      },
      servers: [
        {
          url: "http://localhost:8080",
          description: "Local Server",
        },
      ],
    },
    {
      operationMapper,
      setOperationId: true,
    }
  ),
  {
    components: {
      securitySchemes: {
        bearerAuth: {
          type: "http",
          scheme: "bearer",
          bearerFormat: "JWT",
        },
        "x-service-token": {
          type: "apiKey",
          name: "x-service-token",
          in: "header",
        },
      },
      schemas: {
        ErrorResponse: {
          type: "object",
          required: ["code", "message"],
          properties: {
            code: {
              type: "string",
              description: "Error code",
            },
            message: {
              type: "string",
              description: "Error message",
            },
            details: {
              type: "array",
              items: {
                type: "object",
                required: ["field", "message"],
                properties: {
                  field: {
                    type: "string",
                    description: "Field name",
                  },
                  message: {
                    type: "string",
                    description: "Field error message",
                  },
                },
              },
              description: "Field-level validation errors",
            },
          },
        },
        UnauthorizedResponse: {
          type: "object",
          required: ["code", "message"],
          properties: {
            code: {
              type: "string",
              enum: ["UNAUTHORIZED"],
            },
            message: {
              type: "string",
            },
          },
        },
        RateLimitResponse: {
          type: "object",
          required: ["code", "message"],
          properties: {
            code: {
              type: "string",
              enum: ["RATE_LIMIT_EXCEEDED"],
            },
            message: {
              type: "string",
            },
          },
        },
        NotFoundResponse: {
          type: "object",
          required: ["code", "message"],
          properties: {
            code: {
              type: "string",
              enum: ["NOT_FOUND"],
            },
            message: {
              type: "string",
            },
          },
        },
        ValidationErrorResponse: {
          type: "object",
          required: ["code", "message", "details"],
          properties: {
            code: {
              type: "string",
              enum: ["VALIDATION_ERROR"],
            },
            message: {
              type: "string",
            },
            details: {
              type: "array",
              items: {
                type: "object",
                required: ["field", "message"],
                properties: {
                  field: {
                    type: "string",
                  },
                  message: {
                    type: "string",
                  },
                },
              },
            },
          },
        },
        InternalErrorResponse: {
          type: "object",
          required: ["code", "message"],
          properties: {
            code: {
              type: "string",
              enum: ["INTERNAL_ERROR"],
            },
            message: {
              type: "string",
            },
          },
        },
      },
    },
  }
);
