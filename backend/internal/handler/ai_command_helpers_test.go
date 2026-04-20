package handler

import "testing"

func TestMergePendingCommandKeepsExplicitNewIntent(t *testing.T) {
	pending := PendingAICommand{
		Intent: "create_transaction",
		Data: AICommandTransactionData{
			Title: "Кофе",
		},
	}
	command := &AICommandModelResponse{
		Intent: "create_goal",
		Goal: &AICommandGoalData{
			Title: "Отпуск",
		},
	}

	merged := mergePendingCommand(command, pending)

	if merged.Intent != "create_goal" {
		t.Fatalf("expected create_goal intent, got %s", merged.Intent)
	}
	if merged.Data != nil {
		t.Fatalf("expected transaction data to stay nil for a new explicit intent")
	}
	if merged.Goal == nil || merged.Goal.Title != "Отпуск" {
		t.Fatalf("expected goal payload to remain intact")
	}
}

func TestMergePendingCommandMergesUpdateTransactionDraft(t *testing.T) {
	oldAmount := 2500.0
	newAmount := 3000.0
	pending := PendingAICommand{
		Intent: "update_transaction",
		TransactionSelector: &AICommandTransactionSelector{
			Title: "Такси",
		},
		Data: AICommandTransactionData{
			Type: "expense",
		},
	}
	command := &AICommandModelResponse{
		Intent: "unsupported",
		TransactionSelector: &AICommandTransactionSelector{
			Amount: &oldAmount,
		},
		Data: &AICommandTransactionData{
			Amount: &newAmount,
		},
	}

	merged := mergePendingCommand(command, pending)

	if merged.Intent != "update_transaction" {
		t.Fatalf("expected update_transaction intent, got %s", merged.Intent)
	}
	if merged.TransactionSelector == nil || merged.TransactionSelector.Title != "Такси" {
		t.Fatalf("expected selector title from pending draft")
	}
	if merged.TransactionSelector.Amount == nil || *merged.TransactionSelector.Amount != oldAmount {
		t.Fatalf("expected selector amount from new command")
	}
	if merged.Data == nil || merged.Data.Amount == nil || *merged.Data.Amount != newAmount {
		t.Fatalf("expected updated amount from new command")
	}
	if merged.Data.Type != "expense" {
		t.Fatalf("expected type from pending draft")
	}
}

func TestValidateGoalCommand(t *testing.T) {
	if missing := validateGoalCommand(nil); len(missing) != 2 {
		t.Fatalf("expected two missing fields for nil goal, got %v", missing)
	}

	target := 500000
	deadline := "bad-date"
	missing := validateGoalCommand(&AICommandGoalData{
		Title:       "Отпуск",
		TargetValue: &target,
		Deadline:    &deadline,
	})

	if len(missing) != 1 || missing[0] != "goal.deadline" {
		t.Fatalf("expected invalid deadline, got %v", missing)
	}
}

func TestConfirmationParsing(t *testing.T) {
	if !isPositiveConfirmation("да") {
		t.Fatalf("expected positive confirmation for 'да'")
	}
	if !isPositiveConfirmation("delete") {
		t.Fatalf("expected positive confirmation for 'delete'")
	}
	if !isNegativeConfirmation("нет") {
		t.Fatalf("expected negative confirmation for 'нет'")
	}
	if isNegativeConfirmation("подтверждаю") {
		t.Fatalf("did not expect negative confirmation for 'подтверждаю'")
	}
}
