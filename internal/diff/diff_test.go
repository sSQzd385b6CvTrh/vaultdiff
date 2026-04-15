package diff

import (
	"testing"
)

func TestSecrets_Added(t *testing.T) {
	oldSecrets := map[string]interface{}{"key1": "val1"}
	newSecrets := map[string]interface{}{"key1": "val1", "key2": "val2"}

	result := Secrets("secret/test", oldSecrets, newSecrets)

	if len(result.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(result.Changes))
	}

	var found bool
	for _, c := range result.Changes {
		if c.Key == "key2" && c.Type == Added {
			found = true
		}
	}
	if !found {
		t.Error("expected key2 to be marked as Added")
	}
}

func TestSecrets_Removed(t *testing.T) {
	oldSecrets := map[string]interface{}{"key1": "val1", "key2": "val2"}
	newSecrets := map[string]interface{}{"key1": "val1"}

	result := Secrets("secret/test", oldSecrets, newSecrets)

	var found bool
	for _, c := range result.Changes {
		if c.Key == "key2" && c.Type == Removed {
			found = true
		}
	}
	if !found {
		t.Error("expected key2 to be marked as Removed")
	}
}

func TestSecrets_Modified(t *testing.T) {
	oldSecrets := map[string]interface{}{"key1": "oldval"}
	newSecrets := map[string]interface{}{"key1": "newval"}

	result := Secrets("secret/test", oldSecrets, newSecrets)

	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Type != Modified || c.OldValue != "oldval" || c.NewValue != "newval" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestSecrets_Unchanged(t *testing.T) {
	secrets := map[string]interface{}{"key1": "val1"}
	result := Secrets("secret/test", secrets, secrets)

	if len(result.Changes) != 1 || result.Changes[0].Type != Unchanged {
		t.Error("expected key1 to be Unchanged")
	}
}

func TestSecrets_SortedOutput(t *testing.T) {
	oldSecrets := map[string]interface{}{}
	newSecrets := map[string]interface{}{"zebra": "z", "apple": "a", "mango": "m"}

	result := Secrets("secret/test", oldSecrets, newSecrets)

	keys := []string{result.Changes[0].Key, result.Changes[1].Key, result.Changes[2].Key}
	if keys[0] != "apple" || keys[1] != "mango" || keys[2] != "zebra" {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}
