"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getWarehouseOrder,
  listLogisticChannels,
  listWarehouseOrders,
  packOrder,
  pickOrder,
  shipOrder,
} from "@/lib/api/warehouse";
import type { WmsStatus } from "@/lib/schemas";

export const warehouseQueryKeys = {
  orders: (status?: WmsStatus) => ["warehouse-orders", status ?? "all"] as const,
  order: (orderSn: string) => ["warehouse-order", orderSn] as const,
  logistics: ["logistic-channels"] as const,
};

export function useWarehouseOrders(status?: WmsStatus) {
  return useQuery({
    queryKey: warehouseQueryKeys.orders(status),
    queryFn: () => listWarehouseOrders(status),
  });
}

export function useWarehouseOrder(orderSn: string | null) {
  return useQuery({
    queryKey: warehouseQueryKeys.order(orderSn ?? ""),
    queryFn: () => getWarehouseOrder(orderSn ?? ""),
    enabled: Boolean(orderSn),
  });
}

export function useLogisticChannels() {
  return useQuery({
    queryKey: warehouseQueryKeys.logistics,
    queryFn: listLogisticChannels,
  });
}

function useOrderMutation<TVariables>(
  mutationFn: (variables: TVariables) => Promise<unknown>,
  orderSnResolver: (variables: TVariables) => string,
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn,
    onSuccess: (_, variables) => {
      const orderSn = orderSnResolver(variables);
      void queryClient.invalidateQueries({ queryKey: ["warehouse-orders"] });
      void queryClient.invalidateQueries({ queryKey: warehouseQueryKeys.order(orderSn) });
    },
  });
}

export function usePickOrderMutation() {
  return useOrderMutation((orderSn: string) => pickOrder(orderSn), (orderSn) => orderSn);
}

export function usePackOrderMutation() {
  return useOrderMutation((orderSn: string) => packOrder(orderSn), (orderSn) => orderSn);
}

export function useShipOrderMutation() {
  return useOrderMutation(
    ({ orderSn, channelId }: { orderSn: string; channelId: string }) => shipOrder(orderSn, channelId),
    ({ orderSn }) => orderSn,
  );
}