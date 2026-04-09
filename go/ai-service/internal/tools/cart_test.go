package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

type fakeCartAPI struct {
	cart    clients.Cart
	item    clients.CartItem
	err     error
	seenJWT string
	seenID  string
	seenQty int
}

func (f *fakeCartAPI) GetCart(ctx context.Context, jwt string) (clients.Cart, error) {
	f.seenJWT = jwt
	return f.cart, f.err
}

func (f *fakeCartAPI) AddToCart(ctx context.Context, jwt, productID string, qty int) (clients.CartItem, error) {
	f.seenJWT = jwt
	f.seenID = productID
	f.seenQty = qty
	return f.item, f.err
}

func TestViewCartTool(t *testing.T) {
	fake := &fakeCartAPI{cart: clients.Cart{
		Items: []clients.CartItem{{ID: "i1", ProductID: "p1", ProductName: "Jacket", ProductPrice: 12999, Quantity: 1}},
		Total: 12999,
	}}
	tool := NewViewCartTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenJWT != "tok" {
		t.Errorf("jwt forwarded = %q", fake.seenJWT)
	}
	m := res.Content.(map[string]any)
	if m["total"].(int) != 12999 {
		t.Errorf("total = %+v", m["total"])
	}
}

func TestViewCartTool_RequiresUserID(t *testing.T) {
	tool := NewViewCartTool(&fakeCartAPI{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddToCartTool_Success(t *testing.T) {
	fake := &fakeCartAPI{item: clients.CartItem{ID: "i1", ProductID: "p1", ProductName: "Jacket", ProductPrice: 12999, Quantity: 2}}
	tool := NewAddToCartTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":2}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenID != "p1" || fake.seenQty != 2 {
		t.Errorf("seen id=%q qty=%d", fake.seenID, fake.seenQty)
	}
	m := res.Content.(map[string]any)
	if m["quantity"].(int) != 2 {
		t.Errorf("content = %+v", m)
	}
}

func TestAddToCartTool_BadArgs(t *testing.T) {
	tool := NewAddToCartTool(&fakeCartAPI{})
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":""}`), "user-1"); err == nil {
		t.Error("expected error for empty product_id")
	}
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":0}`), "user-1"); err == nil {
		t.Error("expected error for zero qty")
	}
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":-1}`), "user-1"); err == nil {
		t.Error("expected error for negative qty")
	}
}
