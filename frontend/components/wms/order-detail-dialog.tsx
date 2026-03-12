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
  if (value.includes("pack") || value.includes("ready")) return "warning" as const;
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
        },
      },
    );
  };

  return (
    <UiDialog
      open={isOpen}
      onClose={closeDialog}
      title={selectedOrderSn ? `Order ${selectedOrderSn}` : "Order Detail"}
      className="max-w-6xl border-[#d9def5] bg-[#f7f8ff]"
    >
      {isLoading || !order ? (
        <p className="py-8 text-sm text-[#646c8a]">Loading detail...</p>
      ) : (
        <div className="space-y-5">
          <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-xs uppercase tracking-[0.16em] text-[#7f89b2]">Order Summary</p>
                <h2 className="mt-1 font-[var(--font-space-grotesk)] text-2xl font-semibold text-[#222b4a]">
                  {order.order_sn}
                </h2>
                <p className="mt-1 text-sm text-[#6d7698]">Updated {formatDateTime(order.updated_at)}</p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant={getStatusTone(order.wms_status)}>{order.wms_status.replaceAll("_", " ")}</Badge>
                <Badge variant="outline">{order.marketplace_status ?? "-"}</Badge>
                <Badge variant="outline">{order.shipping_status ?? "-"}</Badge>
              </div>
            </div>
          </section>

          <div className="grid gap-5 lg:grid-cols-[1.45fr_1fr]">
            <div className="space-y-5">
              <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 font-[var(--font-space-grotesk)] text-lg font-semibold text-[#2a3353]">Order Information</h3>
                <div className="grid gap-3 text-sm sm:grid-cols-2">
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Shop ID</p>
                    <p className="mt-1 font-medium text-[#2d3658]">{order.shop_id ?? "-"}</p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Tracking Number</p>
                    <p className="mt-1 font-medium text-[#2d3658]">{order.tracking_number ?? "-"}</p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Created At</p>
                    <p className="mt-1 font-medium text-[#2d3658]">
                      {order.created_at ? formatDateTime(order.created_at) : "-"}
                    </p>
                  </div>
                  <div className="rounded-lg border border-[#edf0fb] bg-[#fafbff] p-3">
                    <p className="text-xs text-[#7f89b2]">Total Amount</p>
                    <p className="mt-1 font-medium text-[#2d3658]">{formatMoney(order.total_amount ?? 0)}</p>
                  </div>
                </div>
              </section>

              <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 font-[var(--font-space-grotesk)] text-lg font-semibold text-[#2a3353]">Order Items</h3>
                {order.items.length ? (
                  <div className="overflow-hidden rounded-lg border border-[#e7eaf8]">
                    <table className="min-w-full text-sm">
                      <thead className="bg-[#f6f8ff]">
                        <tr>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">SKU</th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">Qty</th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">Unit Price</th>
                          <th className="px-3 py-2.5 text-left font-medium text-[#5e678d]">Subtotal</th>
                        </tr>
                      </thead>
                      <tbody>
                        {order.items.map((item, index) => (
                          <tr
                            key={item.sku}
                            className={`border-t border-[#eceff9] ${index % 2 === 0 ? "bg-white" : "bg-[#fbfcff]"}`}
                          >
                            <td className="px-3 py-2.5 text-[#2f3860]">{item.sku}</td>
                            <td className="px-3 py-2.5 text-[#2f3860]">{item.quantity}</td>
                            <td className="px-3 py-2.5 text-[#2f3860]">{formatMoney(item.price)}</td>
                            <td className="px-3 py-2.5 text-[#2f3860]">{formatMoney(item.price * item.quantity)}</td>
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

            <aside className="space-y-5">
              <section className="animate-rise-fade rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 font-[var(--font-space-grotesk)] text-lg font-semibold text-[#2a3353]">Workflow</h3>
                <div className="space-y-2.5 text-sm">
                  <div className="flex items-center gap-2 rounded-md bg-[#f7f9ff] p-2.5 text-[#4f5a83]">
                    <Clock3 className="h-4 w-4" />
                    Ready to Pick
                  </div>
                  <div className="flex items-center gap-2 rounded-md bg-[#f7f9ff] p-2.5 text-[#4f5a83]">
                    <Package className="h-4 w-4" />
                    Picking & Packing
                  </div>
                  <div className="flex items-center gap-2 rounded-md bg-[#f7f9ff] p-2.5 text-[#4f5a83]">
                    <Truck className="h-4 w-4" />
                    Shipping
                  </div>
                  <div className="flex items-center gap-2 rounded-md bg-[#f7f9ff] p-2.5 text-[#4f5a83]">
                    <CircleCheck className="h-4 w-4" />
                    Completed
                  </div>
                </div>
              </section>

              <section className="rounded-xl border border-[#e3e7f8] bg-white p-4">
                <h3 className="mb-3 font-[var(--font-space-grotesk)] text-lg font-semibold text-[#2a3353]">Actions</h3>
                <div className="space-y-3">
                  <Button
                    className="h-10 w-full rounded-lg"
                    variant="outline"
                    disabled={!canPick(order.wms_status) || pickMutation.isPending}
                    onClick={() => pickMutation.mutate(order.order_sn)}
                  >
                    Mark as Picking
                  </Button>
                  <Button
                    className="h-10 w-full rounded-lg"
                    variant="outline"
                    disabled={!canPack(order.wms_status) || packMutation.isPending}
                    onClick={() => packMutation.mutate(order.order_sn)}
                  >
                    Mark as Packed
                  </Button>

                  <form className="space-y-2" onSubmit={form.handleSubmit(onShipSubmit)}>
                    <Select
                      className="h-10 border-[#dce1f6]"
                      disabled={!canShip(order.wms_status)}
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
                      <p className="text-xs text-[#d74747]">{form.formState.errors.channel_id.message}</p>
                    ) : null}
                    <Button
                      type="submit"
                      className="h-10 w-full rounded-lg bg-[#2f66ff] text-white hover:bg-[#1f54e6]"
                      disabled={!canShip(order.wms_status) || shipMutation.isPending}
                    >
                      Ship Order
                    </Button>
                  </form>

                  {pickMutation.error ? <p className="text-xs text-[#d74747]">{pickMutation.error.message}</p> : null}
                  {packMutation.error ? <p className="text-xs text-[#d74747]">{packMutation.error.message}</p> : null}
                  {shipMutation.error ? <p className="text-xs text-[#d74747]">{shipMutation.error.message}</p> : null}
                </div>
              </section>
            </aside>
          </div>
        </div>
      )}
    </UiDialog>
  );
}