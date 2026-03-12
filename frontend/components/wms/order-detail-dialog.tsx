"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { CircleCheck, Clock3, Package, Truck } from "lucide-react";
import { useForm } from "react-hook-form";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { UiDialog } from "@/components/ui/dialog";
import { Select } from "@/components/ui/select";
import {
  useLogisticChannels,
  usePackOrderMutation,
  usePickOrderMutation,
  useShipOrderMutation,
  useWarehouseOrder,
} from "@/hooks/use-warehouse";
import { shipOrderFormSchema, type ShipOrderFormValues } from "@/lib/schemas";
import { useWmsStore } from "@/lib/store/wms-store";
import { formatDateTime, formatMoney } from "@/lib/utils";

function getStatusTone(status: string) {
  const value = status.toLowerCase();
  if (value.includes("ship")) return "success" as const;
  if (value.includes("pack") || value.includes("ready"))
    return "warning" as const;
  return "outline" as const;
}

function canPick(status: string) {
  return status === "READY_TO_PICK";
}

function canPack(status: string) {
  return status === "PICKING";
}

function canShip(status: string) {
  return status === "PACKED";
}

function getDisplayWmsStatus(status: string) {
  if (status === "READY_TO_PICK") return "READY TO PICK";
  if (status === "PICKING") return "PICKING";
  if (status === "PACKED") return "READY TO SHIP";
  if (status === "SHIPPED") return "SHIPPING";
  return status.replaceAll("_", " ");
}

function getWorkflowIndex(status: string) {
  if (status === "READY_TO_PICK") return 0;
  if (status === "PICKING") return 1;
  if (status === "PACKED") return 2;
  if (status === "SHIPPED") return 3;
  return 0;
}

function getActionLabel(status: string) {
  if (status === "READY_TO_PICK") return "Start Picking";
  if (status === "PICKING") return "Complete Picking";
  if (status === "PACKED") return "Start Shipping";
  if (status === "SHIPPED") return "Complete Shipping";
  return "Update Status";
}

export function OrderDetailDialog() {
  const isOpen = useWmsStore((state) => state.detailOpen);

  const selectedOrderSn = useWmsStore((state) => state.selectedOrderSn);
  const closeDialog = useWmsStore((state) => state.closeOrderDetail);

  const { data: order, isLoading } = useWarehouseOrder(selectedOrderSn);
  const { data: logisticChannels = [] } = useLogisticChannels();

  const pickMutation = usePickOrderMutation();
  const packMutation = usePackOrderMutation();
  const shipMutation = useShipOrderMutation();

  const form = useForm<ShipOrderFormValues>({
    resolver: zodResolver(shipOrderFormSchema),
    defaultValues: { channel_id: "" },
  });

  const onShipSubmit = (values: ShipOrderFormValues) => {
    if (!selectedOrderSn) return;
    shipMutation.mutate(
      { orderSn: selectedOrderSn, channelId: values.channel_id },
      {
        onSuccess: () => {
          form.reset({ channel_id: "" });
          closeDialog();
        },
      },
    );
  };

  const handlePickup = (orderSn: string) => {
    pickMutation.mutate(orderSn, {
      onSuccess: () => {
        closeDialog();
      },
    });
  };

  const handlePack = (orderSn: string) => {
    packMutation.mutate(orderSn, {
      onSuccess: () => {
        closeDialog();
      },
    });
  };
  const workflowSteps = [
    { label: "Ready to Pick", icon: Clock3 },
    { label: "Picking", icon: Package },
    { label: "Ready to Ship", icon: Truck },
    { label: "Shipping", icon: CircleCheck },
  ];

  return (
    <UiDialog
      open={isOpen}
      onClose={closeDialog}
      title={selectedOrderSn ? `Order ${selectedOrderSn}` : "Order Detail"}
      className="max-w-200 border-[#d9def5] bg-[#f7f8ff]"
    >
      {isLoading || !order ? (
        <p className="py-8 text-sm text-[#646c8a]">Loading detail...</p>
      ) : (
        <div className="space-y-5 ">
          <section className="rounded-2xl  border border-[#dae2ff] bg-linear-to-r from-[#f7f9ff] via-[#ffffff] to-[#edf3ff] p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-xs uppercase tracking-[0.16em] text-[#7f89b2]">
                  Order Summary
                </p>
                <h2 className="mt-1 text-2xl font-semibold text-[#222b4a]">
                  {order.order_sn}
                </h2>
                <p className="mt-1 text-sm text-[#6d7698]">
                  Updated {formatDateTime(order.updated_at)}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant={getStatusTone(order.wms_status)}>
                  {getDisplayWmsStatus(order.wms_status)}
                </Badge>
                <Badge variant="outline">
                  {order.marketplace_status ?? "-"}
                </Badge>
                <Badge variant="outline">{order.shipping_status ?? "-"}</Badge>
              </div>
            </div>
          </section>

          <div className="flex flex-col gap-2">
            <div className="space-y-5">
              <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 text-lg font-semibold text-[#2a3353]">
                  Order Information
                </h3>
                <div className="grid gap-3 text-sm sm:grid-cols-2">
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Shop ID</p>
                    <p className="mt-1 font-medium text-[#2d3658]">
                      {order.shop_id ?? "-"}
                    </p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Tracking Number</p>
                    <p className="mt-1 font-medium text-[#2d3658]">
                      {order.tracking_number ?? "-"}
                    </p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Created At</p>
                    <p className="mt-1 font-medium text-[#2d3658]">
                      {order.created_at
                        ? formatDateTime(order.created_at)
                        : "-"}
                    </p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Total Amount</p>
                    <p className="mt-1 font-medium text-[#2d3658]">
                      {formatMoney(order.total_amount ?? 0)}
                    </p>
                  </div>
                </div>
              </section>

              <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 text-lg font-semibold text-[#2a3353]">
                  Order Items
                </h3>
                {order.items.length ? (
                  <div className="overflow-hidden rounded-lg border border-[#e7eaf8]">
                    <table className="min-w-full text-sm">
                      <thead className="bg-[#f6f8ff]">
                        <tr>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">
                            SKU
                          </th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">
                            Qty
                          </th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">
                            Unit Price
                          </th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">
                            Subtotal
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {order.items.map((item, index) => (
                          <tr
                            key={item.sku}
                            className={`border-t border-[#eceff9] ${index % 2 === 0 ? "bg-white" : "bg-[#fbfcff]"}`}
                          >
                            <td className="px-3 py-2.5 text-[#2f3860]">
                              {item.sku}
                            </td>
                            <td className="px-3 py-2.5 text-[#2f3860]">
                              {item.quantity}
                            </td>
                            <td className="px-3 py-2.5 text-[#2f3860]">
                              {formatMoney(item.price)}
                            </td>
                            <td className="px-3 py-2.5 text-[#2f3860]">
                              {formatMoney(item.price * item.quantity)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p className="text-sm text-[#6d7698]">No items found.</p>
                )}
              </section>
            </div>

            <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
              <h3 className="mb-3 text-lg font-semibold text-[#2a3353]">
                Actions
              </h3>
              <div className="space-y-3">
                {canPick(order.wms_status) ? (
                  <Button
                    className="h-10 w-full rounded-lg bg-[#2f66ff] text-white hover:bg-[#1f54e6]"
                    disabled={pickMutation.isPending}
                    onClick={() => handlePickup(order.order_sn)}
                  >
                    {getActionLabel(order.wms_status)}
                  </Button>
                ) : null}

                {canPack(order.wms_status) ? (
                  <Button
                    className="h-10 w-full rounded-lg bg-[#2f66ff] text-white hover:bg-[#1f54e6]"
                    disabled={packMutation.isPending}
                    onClick={() => handlePack(order.order_sn)}
                  >
                    {getActionLabel(order.wms_status)}
                  </Button>
                ) : null}

                {canShip(order.wms_status) ? (
                  <form
                    className="space-y-2"
                    onSubmit={form.handleSubmit(onShipSubmit)}
                  >
                    <Select
                      className="h-10 border-[#dce1f6]"
                      disabled={shipMutation.isPending}
                      {...form.register("channel_id")}
                    >
                      <option value="">Select logistic channel</option>
                      {logisticChannels.map((channel) => (
                        <option key={channel.id} value={channel.id}>
                          {channel.name}
                        </option>
                      ))}
                    </Select>
                    {form.formState.errors.channel_id ? (
                      <p className="text-xs text-[#d74747]">
                        {form.formState.errors.channel_id.message}
                      </p>
                    ) : null}
                    <Button
                      type="submit"
                      className="h-10 w-full rounded-lg bg-[#2f66ff] text-white hover:bg-[#1f54e6]"
                      disabled={shipMutation.isPending}
                    >
                      {getActionLabel(order.wms_status)}
                    </Button>
                  </form>
                ) : null}

                {order.wms_status === "SHIPPED" ? (
                  <>
                    <Button
                      className="h-10 w-full rounded-lg"
                      variant="outline"
                      disabled
                    >
                      {getActionLabel(order.wms_status)}
                    </Button>
                    <p className="text-xs text-[#6d7698]">
                      Shipping completion is synced from marketplace updates.
                    </p>
                  </>
                ) : null}

                {pickMutation.error ? (
                  <p className="text-xs text-[#d74747]">
                    {pickMutation.error.message}
                  </p>
                ) : null}
                {packMutation.error ? (
                  <p className="text-xs text-[#d74747]">
                    {packMutation.error.message}
                  </p>
                ) : null}
                {shipMutation.error ? (
                  <p className="text-xs text-[#d74747]">
                    {shipMutation.error.message}
                  </p>
                ) : null}
              </div>
            </section>
          </div>
        </div>
      )}
    </UiDialog>
  );
}
