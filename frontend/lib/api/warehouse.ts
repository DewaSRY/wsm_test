import { z } from "zod";
import { apiClient, unwrapApiData } from "@/lib/api/client";
import {
  logisticChannelsResponseSchema,
  orderDetailSchema,
  orderTransitionResponseSchema,
  warehouseOrdersResponseSchema,
  wmsStatusSchema,
  type OrderDetail,
  type OrderTransitionResponse,
  type WmsStatus,
} from "@/lib/schemas";

const wrappedOrderDetailSchema = z.object({
  order: orderDetailSchema,
});

export async function listWarehouseOrders(status?: WmsStatus) {
  const parsedStatus = status ? wmsStatusSchema.parse(status) : undefined;
  const response = await apiClient.get("/api/warehouse/orders", {
    params: parsedStatus ? { wms_status: parsedStatus } : undefined,
  });
  return warehouseOrdersResponseSchema.parse(response.data);
}

export async function getWarehouseOrder(orderSn: string): Promise<OrderDetail> {
  const response = await apiClient.get(`/api/warehouse/orders/${orderSn}`);
  const payload = unwrapApiData<unknown>(response.data);

  if (payload && typeof payload === "object" && "order" in payload) {
    return wrappedOrderDetailSchema.parse(payload).order;
  }

  return orderDetailSchema.parse(payload);
}

export async function pickOrder(orderSn: string): Promise<OrderTransitionResponse> {
  const response = await apiClient.post(`/api/warehouse/orders/${orderSn}/pick`);
  return orderTransitionResponseSchema.parse(unwrapApiData(response.data));
}

export async function packOrder(orderSn: string): Promise<OrderTransitionResponse> {
  const response = await apiClient.post(`/api/warehouse/orders/${orderSn}/pack`);
  return orderTransitionResponseSchema.parse(unwrapApiData(response.data));
}

export async function shipOrder(orderSn: string, channelId: string): Promise<OrderTransitionResponse> {
  const response = await apiClient.post(`/api/warehouse/orders/${orderSn}/ship`, {
    channel_id: channelId,
  });
  return orderTransitionResponseSchema.parse(unwrapApiData(response.data));
}

export async function listLogisticChannels() {
  const response = await apiClient.get("/api/logistic/channels");
  return logisticChannelsResponseSchema.parse(response.data).data;
}