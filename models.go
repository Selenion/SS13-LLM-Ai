package main

// --- LLM Structures ---

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// --- Incoming Data ---

// Факс от игрока
type FaxRequest struct {
	Sender  string `json:"sender"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// --- Admin Analysis (Step 1) ---

type FaxAnalysisResponse struct {
	Summary   string `json:"summary"`
	Urgency   string `json:"urgency"` // Low/Medium/High
	LatencyMs int64  `json:"latency_ms"`
}

// --- Admin Decision (Step 2) ---

type FaxReplyRequest struct {
	OriginalFax FaxRequest `json:"original_fax"` // Контекст
	Action      string     `json:"action"`       // "approve", "deny", "custom"
	CustomNote  string     `json:"custom_note"`  // Текст админа (только для custom)
}

type FaxReplyResponse struct {
	Draft     string `json:"draft"`
	LatencyMs int64  `json:"latency_ms"`
}
