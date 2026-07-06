package models

import "testing"

func TestOrderStatusValidation(t *testing.T) {
	validStatuses := []string{
		OrderStatusAwaitingPayment,
		OrderStatusPreparation,
		OrderStatusShipping,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}
	for _, status := range validStatuses {
		if !IsValidOrderStatus(status) {
			t.Fatalf("expected order status %q to be valid", status)
		}
	}
	if IsValidOrderStatus("unknown") {
		t.Fatal("expected unknown order status to be invalid")
	}
}

func TestOrderPaymentStatusValidation(t *testing.T) {
	validStatuses := []string{
		OrderPaymentStatusPending,
		OrderPaymentStatusCheckoutOpen,
		OrderPaymentStatusPaid,
		OrderPaymentStatusFailed,
		OrderPaymentStatusExpired,
		OrderPaymentStatusNoPaymentNeeded,
	}
	for _, status := range validStatuses {
		if !IsValidOrderPaymentStatus(status) {
			t.Fatalf("expected payment status %q to be valid", status)
		}
	}
	if IsValidOrderPaymentStatus("unknown") {
		t.Fatal("expected unknown payment status to be invalid")
	}
}
