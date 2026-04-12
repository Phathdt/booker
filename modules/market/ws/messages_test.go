package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubscribeMsg_JSON_Marshal tests SubscribeMsg marshaling
func TestSubscribeMsg_JSON_Marshal(t *testing.T) {
	msg := SubscribeMsg{
		Op:      "subscribe",
		Channel: "ticker",
		Pair:    "BTC_USDT",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	// Verify it can be unmarshaled back
	var decoded SubscribeMsg
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, msg, decoded)
}

// TestSubscribeMsg_Unmarshal_FromJSON tests SubscribeMsg unmarshaling
func TestSubscribeMsg_Unmarshal_FromJSON(t *testing.T) {
	jsonStr := `{"op":"unsubscribe","channel":"trades","pair":"ETH_USDT"}`

	var msg SubscribeMsg
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, "unsubscribe", msg.Op)
	assert.Equal(t, "trades", msg.Channel)
	assert.Equal(t, "ETH_USDT", msg.Pair)
}

// TestSubscribeMsg_AllFields tests all SubscribeMsg fields
func TestSubscribeMsg_AllFields(t *testing.T) {
	msg := SubscribeMsg{
		Op:      "subscribe",
		Channel: "ticker",
		Pair:    "BTC_USDT",
	}

	assert.Equal(t, "subscribe", msg.Op)
	assert.Equal(t, "ticker", msg.Channel)
	assert.Equal(t, "BTC_USDT", msg.Pair)
}

// TestWSMessage_JSON_Marshal tests WSMessage marshaling
func TestWSMessage_JSON_Marshal(t *testing.T) {
	msg := WSMessage{
		Type: "ticker",
		Pair: "BTC_USDT",
		Data: map[string]string{"price": "45000"},
		Msg:  "",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded WSMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "ticker", decoded.Type)
	assert.Equal(t, "BTC_USDT", decoded.Pair)
}

// TestWSMessage_WithError tests WSMessage for error type
func TestWSMessage_WithError(t *testing.T) {
	msg := WSMessage{
		Type: "error",
		Pair: "",
		Data: nil,
		Msg:  "invalid message format",
	}

	assert.Equal(t, "error", msg.Type)
	assert.Equal(t, "", msg.Pair)
	assert.Nil(t, msg.Data)
	assert.Equal(t, "invalid message format", msg.Msg)
}

// TestWSMessage_OmitEmptyFields tests json omitempty behavior
func TestWSMessage_OmitEmptyFields(t *testing.T) {
	msg := WSMessage{
		Type: "error",
		Pair: "",  // Should be omitted
		Data: nil, // Should be omitted
		Msg:  "error",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	// Check that empty fields are not in JSON
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"type":"error"`)
	assert.Contains(t, jsonStr, `"msg":"error"`)
	// Empty pair should still be there but omitted in fields with omitempty
}

// TestTickerData_JSON_Marshal tests TickerData marshaling
func TestTickerData_JSON_Marshal(t *testing.T) {
	data := TickerData{
		Open:      "44000.00",
		High:      "46000.00",
		Low:       "43000.00",
		Close:     "45000.00",
		Volume:    "1000.00",
		ChangePct: "2.27",
		LastPrice: "45000.00",
		Timestamp: 1234567890,
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var decoded TickerData
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

// TestTickerData_AllFields tests all TickerData fields
func TestTickerData_AllFields(t *testing.T) {
	ticker := TickerData{
		Open:      "40000.00",
		High:      "50000.00",
		Low:       "39000.00",
		Close:     "45500.00",
		Volume:    "5000.50",
		ChangePct: "5.35",
		LastPrice: "45500.00",
		Timestamp: 1609459200,
	}

	assert.Equal(t, "40000.00", ticker.Open)
	assert.Equal(t, "50000.00", ticker.High)
	assert.Equal(t, "39000.00", ticker.Low)
	assert.Equal(t, "45500.00", ticker.Close)
	assert.Equal(t, "5000.50", ticker.Volume)
	assert.Equal(t, "5.35", ticker.ChangePct)
	assert.Equal(t, "45500.00", ticker.LastPrice)
	assert.Equal(t, int64(1609459200), ticker.Timestamp)
}

// TestTradeData_JSON_Marshal tests TradeData marshaling
func TestTradeData_JSON_Marshal(t *testing.T) {
	trade := TradeData{
		TradeID:   "trade-12345",
		Price:     "45000.00",
		Quantity:  "0.5",
		Timestamp: 1234567890,
	}

	data, err := json.Marshal(trade)
	require.NoError(t, err)

	var decoded TradeData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, trade, decoded)
}

// TestTradeData_AllFields tests all TradeData fields
func TestTradeData_AllFields(t *testing.T) {
	trade := TradeData{
		TradeID:   "trade-999",
		Price:     "50000.00",
		Quantity:  "1.5",
		Timestamp: 1609459200,
	}

	assert.Equal(t, "trade-999", trade.TradeID)
	assert.Equal(t, "50000.00", trade.Price)
	assert.Equal(t, "1.5", trade.Quantity)
	assert.Equal(t, int64(1609459200), trade.Timestamp)
}

// TestWSMessage_WithTickerData tests WSMessage containing TickerData
func TestWSMessage_WithTickerData(t *testing.T) {
	ticker := TickerData{
		Open:      "44000.00",
		High:      "46000.00",
		Low:       "43000.00",
		Close:     "45000.00",
		Volume:    "1000.00",
		ChangePct: "2.27",
		LastPrice: "45000.00",
		Timestamp: 1234567890,
	}

	msg := WSMessage{
		Type: "ticker",
		Pair: "BTC_USDT",
		Data: ticker,
		Msg:  "",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded WSMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "ticker", decoded.Type)
	assert.Equal(t, "BTC_USDT", decoded.Pair)
}

// TestWSMessage_WithTradeData tests WSMessage containing TradeData
func TestWSMessage_WithTradeData(t *testing.T) {
	trade := TradeData{
		TradeID:   "trade-123",
		Price:     "45000.00",
		Quantity:  "0.5",
		Timestamp: 1234567890,
	}

	msg := WSMessage{
		Type: "trade",
		Pair: "BTC_USDT",
		Data: trade,
		Msg:  "",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded WSMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "trade", decoded.Type)
	assert.Equal(t, "BTC_USDT", decoded.Pair)
}

// TestSubscribeMsg_UnsubscribeOp tests unsubscribe operation
func TestSubscribeMsg_UnsubscribeOp(t *testing.T) {
	msg := SubscribeMsg{
		Op:      "unsubscribe",
		Channel: "ticker",
		Pair:    "ETH_USDT",
	}

	assert.Equal(t, "unsubscribe", msg.Op)
}

// TestTickerData_WithZeroValues tests TickerData with zero values
func TestTickerData_WithZeroValues(t *testing.T) {
	ticker := TickerData{
		Open:      "0",
		High:      "0",
		Low:       "0",
		Close:     "0",
		Volume:    "0",
		ChangePct: "0",
		LastPrice: "0",
		Timestamp: 0,
	}

	data, err := json.Marshal(ticker)
	require.NoError(t, err)

	var decoded TickerData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, ticker, decoded)
}

// TestTradeData_WithZeroValues tests TradeData with zero values
func TestTradeData_WithZeroValues(t *testing.T) {
	trade := TradeData{
		TradeID:   "",
		Price:     "0",
		Quantity:  "0",
		Timestamp: 0,
	}

	data, err := json.Marshal(trade)
	require.NoError(t, err)

	var decoded TradeData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, trade, decoded)
}

// TestSubscribeMsg_EmptyValues tests SubscribeMsg with empty values
func TestSubscribeMsg_EmptyValues(t *testing.T) {
	msg := SubscribeMsg{
		Op:      "",
		Channel: "",
		Pair:    "",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded SubscribeMsg
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, msg, decoded)
}

// TestWSMessage_Unmarshal_MissingFields tests unmarshaling with missing fields
func TestWSMessage_Unmarshal_MissingFields(t *testing.T) {
	jsonStr := `{"type":"ticker"}`

	var msg WSMessage
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, "ticker", msg.Type)
	assert.Equal(t, "", msg.Pair)
	assert.Nil(t, msg.Data)
	assert.Equal(t, "", msg.Msg)
}

// TestTickerData_LargeNumbers tests TickerData with large numbers
func TestTickerData_LargeNumbers(t *testing.T) {
	ticker := TickerData{
		Open:      "999999999999.99",
		High:      "999999999999.99",
		Low:       "999999999999.99",
		Close:     "999999999999.99",
		Volume:    "999999999999.99",
		ChangePct: "999999999999.99",
		LastPrice: "999999999999.99",
		Timestamp: 9223372036854775807, // Max int64
	}

	data, err := json.Marshal(ticker)
	require.NoError(t, err)

	var decoded TickerData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, ticker, decoded)
}

// TestTradeData_SpecialCharacters tests TradeData with special characters
func TestTradeData_SpecialCharacters(t *testing.T) {
	trade := TradeData{
		TradeID:   "trade-123-@#$%^&*()",
		Price:     "123.45",
		Quantity:  "10.5",
		Timestamp: 1234567890,
	}

	data, err := json.Marshal(trade)
	require.NoError(t, err)

	var decoded TradeData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, trade, decoded)
}
