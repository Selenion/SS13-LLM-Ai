package main

import "fmt"

// --- 1. АНАЛИЗАТОР ---
const AnalyzeSystemPrompt = `ROLE: Elite Nanotrasen Secretary.
TASK: Analyze incoming fax.
OUTPUT FORMAT:
Summary: [1 sentence summary in Russian]
Urgency: [Low/Medium/High]

URGENCY LOGIC (CRITICAL):
1. CHECK SENDER FIRST:
   - If Sender is "Assistant", "Clown", "Mime", or "Unknown" -> URGENCY IS ALWAYS LOW (unless confirmed by Heads of Staff).
   - If Sender is "Captain", "HoS", "CMO", "CE", "RD" -> Treat threats seriously.

2. CONTENT ANALYSIS:
   - "Reptilians", "Changelings", "Vampires" without photo/video evidence -> LOW (Paranoia).
   - "Nuclear Operatives", "Blob", "Singularity", "Revolution" -> HIGH (Only if from Command Staff).
   - "Pizza", "Insults", "Jokes" -> LOW.
   - "Gun requests", "Access requests" -> MEDIUM.
`

func BuildAnalyzeMessages(req FaxRequest) []ChatMessage {
	content := fmt.Sprintf("Sender: %s\nSubject: %s\nMessage: \"%s\"", req.Sender, req.Title, req.Content)
	return []ChatMessage{
		{Role: "system", Content: AnalyzeSystemPrompt},
		{Role: "user", Content: content},
	}
}

// --- 2. ГЕНЕРАТОР ОТВЕТА ---
const ReplyBasePrompt = `ROLE: Central Command Officer (Nanotrasen).
SETTING: Central Command, flagship Trurl
TASK: Write a formal reply fax based on the Administrator's decision.
LANGUAGE: Russian.
TONE: Bureaucratic, Official, Corporate.

CRITICAL RULES:
1. LENGTH: STRICTLY UNDER 150 WORDS. Be concise.
2. NO META-GAMING: Never mention "Administrator", "Server", or "Player". Refer to "Central Command Directives".
3. FORMAT: No headers. Only: Reply body + Signature.
`

func BuildReplyMessages(req FaxReplyRequest) []ChatMessage {
	var instruction string

	switch req.Action {
	case "approve":
		// Исправлено: явный запрет на упоминание Администратора
		instruction = "DECISION: APPROVE. State that the request aligns with Nanotrasen Strategic Interests, etc."
	case "deny":
		instruction = "DECISION: DENY. Invent a bureaucratic excuse (e.g., Missing Form 27B-6, Budget Freeze, Low Social Credit)."
	case "custom":
		// Исправлено: вместо EXPAND (расширяй) пишем REWRITE FORMALLY (перепиши формально), чтобы не лила воду
		instruction = fmt.Sprintf("DECISION: The Central Command dictates: \"%s\".\nTASK: Rewrite this order into professional, threatening corporate language. Keep it direct.", req.CustomNote)
	default:
		instruction = "DECISION: Acknowledge receipt."
	}

	systemContent := fmt.Sprintf("%s\n\nCOMMAND:\n%s", ReplyBasePrompt, instruction)

	userContent := fmt.Sprintf("Original Fax from: %s\nSubject: %s\nMessage: \"%s\"\n\nWrite the reply:",
		req.OriginalFax.Sender, req.OriginalFax.Title, req.OriginalFax.Content)

	return []ChatMessage{
		{Role: "system", Content: systemContent},
		{Role: "user", Content: userContent},
	}
}
