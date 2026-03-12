import { z } from "zod";

export const wmsStatusSchema = z.enum(["READY_TO_PICK", "PICKING", "PACKED", "SHIPPED"]);
export type WmsStatus = z.infer<typeof wmsStatusSchema>;

export const orderListItemSchema = z.object({
  order_sn: z.string(),
  wms_status: wmsStatusSchema,
  marketplace_status: z.string().nullable().optional(),
  shipping_status: z.string().nullable().optional(),
  tracking_number: z.string().nullable().optional(),
  updated_at: z.string(),
});

export const warehouseOrdersResponseSchema = z.object({
  orders: z.array(orderListItemSchema),
});

export const orderItemSchema = z.object({
  sku: z.string(),
  quantity: z.number(),
  price: z.number(),
});

export const orderDetailSchema = orderListItemSchema.extend({
  shop_id: z.union([z.string(), z.number()]).optional(),
  total_amount: z.number().optional(),
  items: z.array(orderItemSchema).default([]),
  created_at: z.string().optional(),
});

export const logisticChannelSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string().optional(),
});

export const logisticChannelsResponseSchema = z.object({
  data: z.array(logisticChannelSchema),
});

export const orderTransitionResponseSchema = z.object({
  order_sn: z.string(),
  wms_status: wmsStatusSchema,
  shipping_status: z.string().nullable().optional(),
  tracking_number: z.string().nullable().optional(),
});

export const shipOrderFormSchema = z.object({
  channel_id: z.string().min(1, "Channel is required"),
});

export type OrderListItem = z.infer<typeof orderListItemSchema>;
export type OrderDetail = z.infer<typeof orderDetailSchema>;
export type LogisticChannel = z.infer<typeof logisticChannelSchema>;
export type OrderTransitionResponse = z.infer<typeof orderTransitionResponseSchema>;
export type ShipOrderFormValues = z.infer<typeof shipOrderFormSchema>;