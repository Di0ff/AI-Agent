package llm

import "strings"

type TaskCategory string

const (
	CategoryNavigation TaskCategory = "navigation"
	CategoryForm       TaskCategory = "form"
	CategoryExtraction TaskCategory = "extraction"
	CategoryPurchase   TaskCategory = "purchase"
	CategoryGeneral    TaskCategory = "general"
)

func DetectTaskCategory(task string) TaskCategory {
	taskLower := strings.ToLower(task)

	navigationKeywords := []string{"открой", "перейди", "найди страницу", "navigate", "go to", "open", "visit"}
	for _, kw := range navigationKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryNavigation
		}
	}

	formKeywords := []string{"заполни", "введи", "отправь форму", "зарегистрируйся", "войди", "fill", "submit", "login", "register"}
	for _, kw := range formKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryForm
		}
	}

	extractionKeywords := []string{"найди", "извлеки", "получи информацию", "скопируй", "прочитай", "find", "extract", "get", "read"}
	for _, kw := range extractionKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryExtraction
		}
	}

	purchaseKeywords := []string{"купи", "оплати", "закажи", "добавь в корзину", "buy", "purchase", "order", "checkout", "cart"}
	for _, kw := range purchaseKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryPurchase
		}
	}

	return CategoryGeneral
}

func GetSystemPromptForCategory(category TaskCategory) string {
	switch category {
	case CategoryNavigation:
		return `You are an expert web navigation agent. You excel at:
- Finding and following links efficiently
- Using search functionality to locate pages
- Understanding site structure and navigation patterns
- Handling redirects and page loads gracefully

Focus on speed and accuracy when navigating to the target page.`

	case CategoryForm:
		return `You are an expert form filling agent. You excel at:
- Identifying all required and optional form fields
- Filling forms accurately with appropriate data
- Handling validation errors and retrying with corrections
- Recognizing different input types (text, select, checkbox, radio)
- Submitting forms only when all required fields are complete

Be thorough and careful with form data. Validate before submitting.`

	case CategoryExtraction:
		return `You are an expert data extraction agent. You excel at:
- Identifying relevant information on pages
- Extracting structured data accurately
- Handling pagination and dynamic content
- Recognizing tables, lists, and other data structures
- Filtering noise and focusing on requested information

Be precise and comprehensive in data extraction.`

	case CategoryPurchase:
		return `You are an expert e-commerce agent. You excel at:
- Finding products and comparing options
- Adding items to cart correctly
- Understanding checkout flows
- Recognizing payment and shipping forms

IMPORTANT: Always ask for user confirmation before:
- Adding expensive items to cart
- Proceeding to payment
- Submitting orders
- Entering payment information

Safety and user control are paramount in purchase operations.`

	default:
		return `You are an AI agent controlling a web browser. You excel at understanding user intent and executing multi-step tasks efficiently and safely.`
	}
}

func GetFewShotExamplesForCategory(category TaskCategory) string {
	switch category {
	case CategoryNavigation:
		return `
Example navigation task:
Task: "Go to the contact page"
Step 1: Look for navigation menu or footer links containing "Contact", "Contact Us", or similar
Step 2: Click the contact link
Step 3: Verify you're on the contact page by checking the page title or URL

Example with search:
Task: "Find the pricing page"
Step 1: Look for "Pricing" in navigation menu
Step 2: If not found, use site search with query "pricing"
Step 3: Click the most relevant result`

	case CategoryForm:
		return `
Example form task:
Task: "Fill out the contact form with name John and email john@example.com"
Step 1: Identify all form fields (name, email, message, etc.)
Step 2: Fill "name" field with "John"
Step 3: Fill "email" field with "john@example.com"
Step 4: Check if any required fields are missing
Step 5: Submit the form only when all required fields are filled`

	case CategoryExtraction:
		return `
Example extraction task:
Task: "Get the list of features from this page"
Step 1: Scan the page for headings like "Features", "What we offer", etc.
Step 2: Identify the list or section containing features
Step 3: Extract each feature item with its description
Step 4: Format the results clearly`

	case CategoryPurchase:
		return `
Example purchase task:
Task: "Add a laptop under $1000 to cart"
Step 1: Search for "laptop"
Step 2: Apply price filter "under $1000"
Step 3: Review results and identify suitable option
Step 4: Click on product to view details
Step 5: Ask user: "Found [Product Name] for $[Price]. Add to cart?"
Step 6: If approved, click "Add to Cart" button`

	default:
		return ""
	}
}

func GetTaskSpecificGuidance(category TaskCategory) string {
	switch category {
	case CategoryNavigation:
		return `
Key considerations:
- Check URL and page title to confirm successful navigation
- Handle loading states and wait for page to be fully loaded
- If direct link not found, try search functionality
- Be aware of mega-menus and nested navigation`

	case CategoryForm:
		return `
Key considerations:
- Always identify ALL fields before starting to fill
- Pay attention to field labels, placeholders, and required markers (*)
- Handle dropdowns, checkboxes, and radio buttons appropriately
- Watch for inline validation messages
- Don't submit until all required fields are correctly filled`

	case CategoryExtraction:
		return `
Key considerations:
- Scroll through entire page to find all relevant information
- Handle paginated content by navigating through pages
- Recognize and handle dynamic content (AJAX loading)
- Structure extracted data clearly and completely`

	case CategoryPurchase:
		return `
Key considerations:
- ALWAYS verify product details before adding to cart
- ALWAYS ask for user confirmation for financial actions
- Never enter payment information without explicit approval
- Be aware of total costs including shipping and taxes
- Recognize "Subscribe" vs "One-time purchase" options`

	default:
		return ""
	}
}
