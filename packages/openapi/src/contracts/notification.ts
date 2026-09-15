import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZNotification,
  ZCreateNotificationRequest,
  ZUnreadCountResponse,
  ZErrorResponse,
  schemaWithPagination,
} from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const notificationContract = c.router({
  listNotifications: {
    summary: "List notifications",
    description: "List notifications for the authenticated user with pagination.",
    path: "/api/v1/user/notifications",
    method: "GET",
    query: z.object({
      page: z.number().int().positive().optional(),
      page_size: z.number().int().min(1).max(100).optional(),
      sort: z.string().optional(),
      order: z.enum(["asc", "desc"]).optional(),
      unread: z.boolean().optional(),
    }),
    responses: {
      200: schemaWithPagination(ZNotification),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getNotification: {
    summary: "Get notification by ID",
    description: "Get a notification by ID.",
    path: "/api/v1/user/notifications/:id",
    method: "GET",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZNotification,
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  createNotification: {
    summary: "Create a notification",
    description: "Create a new notification for the authenticated user.",
    path: "/api/v1/user/notifications",
    method: "POST",
    body: ZCreateNotificationRequest,
    responses: {
      201: ZNotification,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  markNotificationAsRead: {
    summary: "Mark notification as read",
    description: "Mark a single notification as read.",
    path: "/api/v1/user/notifications/:id/read",
    method: "PATCH",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: z.void(),
    responses: {
      204: z.void(),
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  markAllNotificationsAsRead: {
    summary: "Mark all notifications as read",
    description: "Mark all notifications as read for the authenticated user.",
    path: "/api/v1/user/notifications/read-all",
    method: "PATCH",
    body: z.void(),
    responses: {
      204: z.void(),
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  deleteNotification: {
    summary: "Delete notification",
    description: "Delete a notification by ID.",
    path: "/api/v1/user/notifications/:id",
    method: "DELETE",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      204: z.void(),
      401: ZErrorResponse,
      404: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  getUnreadCount: {
    summary: "Get unread notification count",
    description: "Get the count of unread notifications for the authenticated user.",
    path: "/api/v1/user/notifications/unread-count",
    method: "GET",
    responses: {
      200: ZUnreadCountResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
