package extractor

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type ElementInfo struct {
	Tag         string
	Text        string
	Selector    string
	Visible     bool
	Interactive bool
	InViewport  bool
	Bounds      Bounds
	Role        string
	Label       string
	Priority    int
}

type Bounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type PageSnapshot struct {
	URL               string
	Title             string
	Elements          []ElementInfo
	Viewport          Bounds
	AccessibilityTree string
}

func ExtractPageSnapshot(ctx context.Context, page playwright.Page) (*PageSnapshot, error) {
	url := page.URL()

	title, err := page.Title()
	if err != nil {
		title = ""
	}

	viewport, err := getViewport(page)
	if err != nil {
		viewport = Bounds{}
	}

	elements, err := extractElements(ctx, page)
	if err != nil {
		return nil, fmt.Errorf("ошибка извлечения элементов: %w", err)
	}

	accessibilityTree, err := getAccessibilityTree(page)
	if err != nil {
		accessibilityTree = ""
	}

	return &PageSnapshot{
		URL:               url,
		Title:             title,
		Elements:          elements,
		Viewport:          viewport,
		AccessibilityTree: accessibilityTree,
	}, nil
}

func getViewport(page playwright.Page) (Bounds, error) {
	result, err := page.Evaluate(`() => ({
		width: window.innerWidth,
		height: window.innerHeight
	})`)
	if err != nil {
		return Bounds{}, err
	}

	viewportMap, ok := result.(map[string]interface{})
	if !ok {
		return Bounds{}, fmt.Errorf("неверный формат viewport")
	}

	width := 0.0
	height := 0.0

	if w, ok := viewportMap["width"].(float64); ok {
		width = w
	}
	if h, ok := viewportMap["height"].(float64); ok {
		height = h
	}

	return Bounds{
		Width:  width,
		Height: height,
	}, nil
}

func extractElements(ctx context.Context, page playwright.Page) ([]ElementInfo, error) {
	jsCode := `
		() => {
			const elements = [];
			const interactiveSelectors = [
				'button', 'a', 'input', 'select', 'textarea', 
				'[role=button]', '[role=link]', '[role=tab]',
				'[onclick]', '[href]'
			];
			
			const allElements = document.querySelectorAll('*');
			const viewport = {
				top: 0,
				left: 0,
				right: window.innerWidth,
				bottom: window.innerHeight
			};
			
			allElements.forEach(el => {
				const rect = el.getBoundingClientRect();
				const style = window.getComputedStyle(el);
				
				const isVisible = style.display !== 'none' && 
					style.visibility !== 'hidden' && 
					style.opacity !== '0' &&
					rect.width > 0 && 
					rect.height > 0;
				
				if (!isVisible) return;
				
				const isInteractive = interactiveSelectors.some(sel => {
					try {
						return el.matches(sel);
					} catch(e) {
						return false;
					}
				});
				
				const inViewport = rect.top >= viewport.top && 
					rect.left >= viewport.left &&
					rect.bottom <= viewport.bottom &&
					rect.right <= viewport.right;
				
				const text = el.textContent?.trim() || '';
				if (!text && !isInteractive) return;
				
				const selector = buildSelector(el);
				const role = el.getAttribute('role') || '';
				const label = el.getAttribute('aria-label') || 
					el.getAttribute('title') || 
					el.getAttribute('alt') || '';
				
				elements.push({
					tag: el.tagName.toLowerCase(),
					text: text.substring(0, 200),
					selector: selector,
					visible: true,
					interactive: isInteractive,
					inViewport: inViewport,
					bounds: {
						x: rect.x,
						y: rect.y,
						width: rect.width,
						height: rect.height
					},
					role: role,
					label: label,
					priority: calculatePriority(el, isInteractive, inViewport, text)
				});
			});
			
			function buildSelector(el) {
				if (el.id && isUniqueID(el.id)) {
					return '#' + el.id;
				}

				if (el.getAttribute('data-testid')) {
					return '[data-testid="' + el.getAttribute('data-testid') + '"]';
				}

				if (el.name) {
					return '[name="' + el.name + '"]';
				}

				const ariaLabel = el.getAttribute('aria-label');
				const role = el.getAttribute('role');
				if (ariaLabel && role) {
					return '[role="' + role + '"][aria-label="' + ariaLabel + '"]';
				}

				if (el.className) {
					const classes = el.className.split(' ').filter(c => c && !isCommonClass(c));
					if (classes.length > 0) {
						return el.tagName.toLowerCase() + '.' + classes[0];
					}
				}

				const parent = el.parentElement;
				if (parent) {
					const siblings = Array.from(parent.children).filter(child => child.tagName === el.tagName);
					if (siblings.length > 1) {
						const index = siblings.indexOf(el) + 1;
						return el.tagName.toLowerCase() + ':nth-child(' + index + ')';
					}
				}

				return el.tagName.toLowerCase();
			}

			function isUniqueID(id) {
				const commonIDs = ['content', 'main', 'header', 'footer', 'nav', 'menu'];
				return !commonIDs.includes(id.toLowerCase());
			}

			function isCommonClass(className) {
				const commonClasses = ['container', 'wrapper', 'row', 'col', 'btn', 'button',
					'active', 'disabled', 'hidden', 'visible', 'flex', 'grid'];
				const lower = className.toLowerCase();
				return commonClasses.some(c => lower.includes(c));
			}
			
			function calculatePriority(el, isInteractive, inViewport, text) {
				let priority = 1;

				if (isInteractive) priority += 3;
				if (inViewport) priority += 2;

				const role = el.getAttribute('role') || '';
				const ariaLabel = el.getAttribute('aria-label') || '';

				const semanticRoles = ['button', 'link', 'navigation', 'search', 'form', 'textbox', 'menuitem'];
				if (semanticRoles.includes(role.toLowerCase())) {
					priority += 2;
				}

				if (ariaLabel) {
					priority += 1;
				}

				const semanticTags = ['button', 'a', 'input', 'select', 'textarea', 'nav', 'form'];
				if (semanticTags.includes(el.tagName.toLowerCase())) {
					priority += 1;
				}

				return priority;
			}
			
			return elements;
		}
	`

	result, err := page.Evaluate(jsCode)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения JavaScript: %w", err)
	}

	elementsData, ok := result.([]interface{})
	if !ok {
		return []ElementInfo{}, nil
	}

	elements := make([]ElementInfo, 0, len(elementsData))
	for _, elemData := range elementsData {
		elemMap, ok := elemData.(map[string]interface{})
		if !ok {
			continue
		}

		elem := parseElementInfo(elemMap)
		if elem != nil {
			elements = append(elements, *elem)
		}
	}

	return elements, nil
}

func parseElementInfo(data map[string]interface{}) *ElementInfo {
	elem := &ElementInfo{}

	if tag, ok := data["tag"].(string); ok {
		elem.Tag = tag
	}
	if text, ok := data["text"].(string); ok {
		elem.Text = text
	}
	if selector, ok := data["selector"].(string); ok {
		elem.Selector = selector
	}
	if visible, ok := data["visible"].(bool); ok {
		elem.Visible = visible
	}
	if interactive, ok := data["interactive"].(bool); ok {
		elem.Interactive = interactive
	}
	if inViewport, ok := data["inViewport"].(bool); ok {
		elem.InViewport = inViewport
	}
	if role, ok := data["role"].(string); ok {
		elem.Role = role
	}
	if label, ok := data["label"].(string); ok {
		elem.Label = label
	}
	if priority, ok := data["priority"].(float64); ok {
		elem.Priority = int(priority)
	}

	if boundsData, ok := data["bounds"].(map[string]interface{}); ok {
		if x, ok := boundsData["x"].(float64); ok {
			elem.Bounds.X = x
		}
		if y, ok := boundsData["y"].(float64); ok {
			elem.Bounds.Y = y
		}
		if width, ok := boundsData["width"].(float64); ok {
			elem.Bounds.Width = width
		}
		if height, ok := boundsData["height"].(float64); ok {
			elem.Bounds.Height = height
		}
	}

	if elem.Selector == "" {
		return nil
	}

	return elem
}

func getAccessibilityTree(page playwright.Page) (string, error) {
	result, err := page.Evaluate(`
		() => {
			function buildA11yTree(node, depth = 0) {
				if (!node) return '';
				
				const indent = '  '.repeat(depth);
				const role = node.role || '';
				const name = node.name || '';
				const value = node.value || '';
				
				let tree = indent + role;
				if (name) tree += ' "' + name + '"';
				if (value) tree += ' = ' + value;
				tree += '\\n';
				
				if (node.children) {
					node.children.forEach(child => {
						tree += buildA11yTree(child, depth + 1);
					});
				}
				
				return tree;
			}
			
			return new Promise((resolve) => {
				if (window.getComputedStyle) {
					const tree = buildA11yTree({
						role: 'document',
						name: document.title,
						children: Array.from(document.body?.children || []).map(el => ({
							role: el.getAttribute('role') || el.tagName.toLowerCase(),
							name: el.textContent?.trim().substring(0, 50) || '',
							children: []
						}))
					});
					resolve(tree);
				} else {
					resolve('');
				}
			});
		}
	`)
	if err != nil {
		return "", err
	}

	if tree, ok := result.(string); ok {
		return tree, nil
	}

	return "", nil
}
