# Playwright Testing Setup Guide

**Created:** 2026-01-05
**Purpose:** Guide for implementing browser automation testing using Playwright E2E tests and Playwright MCP server.

---

## Overview

This document covers two complementary approaches for browser automation testing:

1. **Playwright E2E Tests** - Traditional automated test suites that run in CI/CD
2. **Playwright MCP Server** - AI-driven interactive browser automation via Model Context Protocol

Both are official Microsoft/Playwright projects and can be used together for comprehensive testing coverage.

---

## Comparison: When to Use Each Approach

| Use Case | Playwright E2E | Playwright MCP |
|----------|---------------|----------------|
| CI/CD pipeline testing | Yes | No |
| Regression testing | Yes | No |
| Interactive exploration | No | Yes |
| AI-assisted testing | No | Yes |
| Keyboard shortcut testing | Yes | Yes |
| Visual verification | Yes | Yes (via snapshots) |
| Reproducible test suites | Yes | No |
| Real-time debugging | Limited | Yes |

### Recommendation

- **Use Playwright MCP** for exploratory testing, debugging, and AI-assisted verification
- **Use Playwright E2E** for regression tests, CI/CD pipelines, and reproducible test suites
- **Use both together**: MCP for initial verification, then convert findings to E2E tests

---

## Part 1: Playwright E2E Tests

### What It Is

Playwright Test is an end-to-end test framework for modern web apps. It:
- Runs tests in Chromium, Firefox, and WebKit
- Supports parallel test execution
- Provides auto-waiting and web-first assertions
- Generates HTML reports with traces
- Can auto-start dev servers before tests

### Installation

```bash
cd ui
npm init playwright@latest
```

When prompted, choose:
- TypeScript (recommended)
- Tests folder: `e2e` (to separate from unit tests in `src/`)
- Add GitHub Actions workflow: Yes (for CI)
- Install Playwright browsers: Yes

This creates:
```
ui/
  playwright.config.ts      # Test configuration
  e2e/
    example.spec.ts         # Example test
```

### Configuration

Create or update `ui/playwright.config.ts`:

```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  // Test directory
  testDir: './e2e',
  
  // Run tests in parallel
  fullyParallel: true,
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter to use
  reporter: [
    ['html', { open: 'never' }],
    ['list']
  ],
  
  // Shared settings for all projects
  use: {
    // Base URL for navigation
    baseURL: 'http://localhost:5173',
    
    // Collect trace when retrying the failed test
    trace: 'on-first-retry',
    
    // Take screenshot on failure
    screenshot: 'only-on-failure',
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    // Optionally add Firefox and WebKit
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Run local dev servers before starting the tests
  webServer: [
    {
      // Start the backend server
      command: '../egenskriven serve',
      url: 'http://localhost:8090/api/health',
      reuseExistingServer: !process.env.CI,
      timeout: 120 * 1000,
    },
    {
      // Start the frontend dev server
      command: 'npm run dev',
      url: 'http://localhost:5173',
      reuseExistingServer: !process.env.CI,
      timeout: 120 * 1000,
    },
  ],
});
```

### Writing Tests

Example test for Phase 6 filter functionality (`ui/e2e/filters.spec.ts`):

```typescript
import { test, expect } from '@playwright/test';

test.describe('Filter System', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Wait for tasks to load
    await page.waitForSelector('[data-testid="task-card"]');
  });

  test('F key opens filter builder', async ({ page }) => {
    // Press F key
    await page.keyboard.press('f');
    
    // Verify filter builder modal opens
    await expect(page.locator('[data-testid="filter-builder"]')).toBeVisible();
  });

  test('Escape closes filter builder', async ({ page }) => {
    // Open filter builder
    await page.keyboard.press('f');
    await expect(page.locator('[data-testid="filter-builder"]')).toBeVisible();
    
    // Press Escape
    await page.keyboard.press('Escape');
    
    // Verify modal closes
    await expect(page.locator('[data-testid="filter-builder"]')).not.toBeVisible();
  });

  test('/ focuses search input', async ({ page }) => {
    // Press / key
    await page.keyboard.press('/');
    
    // Verify search input is focused
    await expect(page.locator('[data-testid="search-input"]')).toBeFocused();
  });

  test('Cmd+B toggles between board and list view', async ({ page }) => {
    // Verify board view is active initially
    await expect(page.locator('[data-testid="board-view"]')).toBeVisible();
    
    // Press Cmd+B (Meta+b on Mac, Control+b on Windows/Linux)
    await page.keyboard.press('Meta+b');
    
    // Verify list view is now active
    await expect(page.locator('[data-testid="list-view"]')).toBeVisible();
    
    // Press Cmd+B again
    await page.keyboard.press('Meta+b');
    
    // Verify board view is back
    await expect(page.locator('[data-testid="board-view"]')).toBeVisible();
  });

  test('filter by priority shows only matching tasks', async ({ page }) => {
    // Open filter builder
    await page.keyboard.press('f');
    
    // Select priority field
    await page.selectOption('[data-testid="filter-field"]', 'priority');
    
    // Select "is" operator
    await page.selectOption('[data-testid="filter-operator"]', 'is');
    
    // Select "high" value
    await page.selectOption('[data-testid="filter-value"]', 'high');
    
    // Add the filter
    await page.click('[data-testid="add-filter-button"]');
    
    // Close modal
    await page.keyboard.press('Escape');
    
    // Verify only high priority tasks are visible
    const tasks = page.locator('[data-testid="task-card"]');
    for (const task of await tasks.all()) {
      await expect(task.locator('[data-testid="priority-badge"]')).toHaveText('high');
    }
  });

  test('search filters tasks in real-time', async ({ page }) => {
    // Get initial task count
    const initialCount = await page.locator('[data-testid="task-card"]').count();
    
    // Focus search and type
    await page.keyboard.press('/');
    await page.keyboard.type('login');
    
    // Wait for debounce (300ms)
    await page.waitForTimeout(400);
    
    // Verify filtered results
    const filteredCount = await page.locator('[data-testid="task-card"]').count();
    expect(filteredCount).toBeLessThan(initialCount);
    
    // Verify all visible tasks contain "login"
    const tasks = page.locator('[data-testid="task-card"]');
    for (const task of await tasks.all()) {
      const title = await task.locator('[data-testid="task-title"]').textContent();
      expect(title?.toLowerCase()).toContain('login');
    }
  });

  test('clear filters shows all tasks', async ({ page }) => {
    // Add a filter first
    await page.keyboard.press('f');
    await page.selectOption('[data-testid="filter-field"]', 'priority');
    await page.selectOption('[data-testid="filter-operator"]', 'is');
    await page.selectOption('[data-testid="filter-value"]', 'high');
    await page.click('[data-testid="add-filter-button"]');
    await page.keyboard.press('Escape');
    
    // Get filtered count
    const filteredCount = await page.locator('[data-testid="task-card"]').count();
    
    // Clear filters
    await page.click('[data-testid="clear-filters-button"]');
    
    // Verify all tasks are visible again
    const allCount = await page.locator('[data-testid="task-card"]').count();
    expect(allCount).toBeGreaterThan(filteredCount);
  });
});

test.describe('Saved Views', () => {
  test('can save current filters as a view', async ({ page }) => {
    await page.goto('/');
    
    // Add a filter
    await page.keyboard.press('f');
    await page.selectOption('[data-testid="filter-field"]', 'priority');
    await page.selectOption('[data-testid="filter-operator"]', 'is');
    await page.selectOption('[data-testid="filter-value"]', 'high');
    await page.click('[data-testid="add-filter-button"]');
    await page.keyboard.press('Escape');
    
    // Save as view
    await page.click('[data-testid="save-view-button"]');
    await page.fill('[data-testid="view-name-input"]', 'High Priority Tasks');
    await page.click('[data-testid="confirm-save-view"]');
    
    // Verify view appears in sidebar
    await expect(page.locator('text=High Priority Tasks')).toBeVisible();
  });
});

test.describe('List View', () => {
  test('displays tasks in table format', async ({ page }) => {
    await page.goto('/');
    
    // Switch to list view
    await page.keyboard.press('Meta+b');
    
    // Verify table structure
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('th:has-text("Status")')).toBeVisible();
    await expect(page.locator('th:has-text("Title")')).toBeVisible();
    await expect(page.locator('th:has-text("Priority")')).toBeVisible();
  });

  test('columns are sortable', async ({ page }) => {
    await page.goto('/');
    await page.keyboard.press('Meta+b');
    
    // Click priority header to sort
    await page.click('th:has-text("Priority")');
    
    // Verify sort indicator
    await expect(page.locator('th:has-text("Priority")')).toContainText('↑');
    
    // Click again to reverse sort
    await page.click('th:has-text("Priority")');
    await expect(page.locator('th:has-text("Priority")')).toContainText('↓');
  });
});
```

### Running Tests

```bash
# Run all tests
npx playwright test

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run a specific test file
npx playwright test e2e/filters.spec.ts

# Run tests in UI mode (interactive)
npx playwright test --ui

# Run tests for a specific project/browser
npx playwright test --project=chromium

# Show HTML report
npx playwright show-report
```

### Adding to package.json

```json
{
  "scripts": {
    "test": "vitest",
    "test:e2e": "playwright test",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:ui": "playwright test --ui"
  }
}
```

### GitHub Actions CI

Create `.github/workflows/e2e-tests.yml`:

```yaml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    timeout-minutes: 60
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build backend
        run: go build -o egenskriven
      
      - name: Install UI dependencies
        run: cd ui && npm ci
      
      - name: Install Playwright Browsers
        run: cd ui && npx playwright install --with-deps
      
      - name: Run Playwright tests
        run: cd ui && npm run test:e2e
      
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: ui/playwright-report/
          retention-days: 30
```

---

## Part 2: Playwright MCP Server

### What It Is

Playwright MCP is an official Microsoft project that provides browser automation capabilities via the Model Context Protocol (MCP). It allows AI agents (like Claude/OpenCode) to:

- Navigate web pages
- Click, type, and interact with elements
- Press keyboard shortcuts
- Take screenshots and snapshots
- Read console messages and network requests
- All without needing vision models (uses accessibility tree)

**GitHub:** https://github.com/microsoft/playwright-mcp
**Stars:** 25.1k
**npm:** `@playwright/mcp`

### Installation for OpenCode

Add to `~/.config/opencode/opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "playwright": {
      "type": "local",
      "command": ["npx", "@playwright/mcp@latest"],
      "enabled": true
    }
  }
}
```

### Configuration Options

The Playwright MCP server accepts many command-line arguments:

```bash
npx @playwright/mcp@latest --help
```

Common options:

| Option | Description | Default |
|--------|-------------|---------|
| `--browser <browser>` | Browser to use: chrome, firefox, webkit, msedge | chrome |
| `--headless` | Run browser in headless mode | headed |
| `--viewport-size <size>` | Viewport size, e.g., "1280x720" | browser default |
| `--user-data-dir <path>` | Path to user data directory for persistence | temp dir |
| `--isolated` | Keep browser profile in memory, don't save to disk | false |
| `--timeout-action <ms>` | Action timeout in milliseconds | 5000 |
| `--timeout-navigation <ms>` | Navigation timeout in milliseconds | 60000 |

Example with options:

```json
{
  "mcp": {
    "playwright": {
      "type": "local",
      "command": [
        "npx", 
        "@playwright/mcp@latest",
        "--browser", "chrome",
        "--viewport-size", "1280x720"
      ],
      "enabled": true
    }
  }
}
```

### Available Tools

The MCP server provides these tools for AI agents:

#### Navigation
- `browser_navigate` - Navigate to a URL
- `browser_navigate_back` - Go back to previous page
- `browser_tabs` - List, create, close, or select tabs

#### Interaction
- `browser_click` - Click on an element
- `browser_type` - Type text into an element
- `browser_fill_form` - Fill multiple form fields
- `browser_select_option` - Select option in dropdown
- `browser_hover` - Hover over an element
- `browser_drag` - Drag and drop between elements

#### Keyboard
- `browser_press_key` - Press a key (e.g., "Escape", "Enter", "f", "Meta+b")

#### State
- `browser_snapshot` - Capture accessibility snapshot (preferred over screenshot)
- `browser_take_screenshot` - Take a screenshot
- `browser_console_messages` - Get console logs
- `browser_network_requests` - List network requests

#### Evaluation
- `browser_evaluate` - Run JavaScript in page context
- `browser_wait_for` - Wait for text or time

#### Lifecycle
- `browser_close` - Close the browser
- `browser_resize` - Resize browser window

### Using Playwright MCP for Testing

Once configured, the AI agent can interactively test the application. Example prompts:

```
"Navigate to http://localhost:5173 and take a snapshot"

"Press the F key and verify the filter builder modal opens"

"Type 'login' in the search input and wait for results to filter"

"Press Meta+b to toggle between board and list view"

"Click on the 'High Priority Tasks' view in the sidebar"

"Fill the filter form with: field=priority, operator=is, value=high"
```

### Accessibility Snapshots vs Screenshots

Playwright MCP uses **accessibility snapshots** by default instead of screenshots:

- **Faster** - No image processing needed
- **Structured** - Returns element tree with roles, names, and refs
- **Deterministic** - Same page = same snapshot
- **LLM-friendly** - Text-based, no vision model needed

Example snapshot output:
```
- main:
  - heading "Task Board" [level=1]
  - button "Filter" [ref=s1e2]
  - textbox "Search..." [ref=s1e3]
  - region "Board":
    - group "Backlog":
      - article "Implement login page" [ref=s1e4]
        - text "high priority"
        - text "feature"
```

The AI can use `ref` values to interact with specific elements.

### Configuration File

For complex setups, use a JSON config file:

```json
{
  "browser": {
    "browserName": "chromium",
    "headless": false,
    "launchOptions": {
      "channel": "chrome"
    },
    "contextOptions": {
      "viewport": { "width": 1280, "height": 720 }
    }
  },
  "server": {
    "port": 8931
  },
  "capabilities": ["core", "pdf", "vision"],
  "timeouts": {
    "action": 5000,
    "navigation": 60000
  }
}
```

Run with:
```bash
npx @playwright/mcp@latest --config path/to/config.json
```

---

## Part 3: Testing Strategy for EgenSkriven

### Recommended Approach

1. **Unit Tests (Vitest)** - Already in place
   - Test individual functions and hooks
   - Mock PocketBase and browser APIs
   - Fast, run on every save

2. **Playwright MCP** - For exploratory/interactive testing
   - AI-assisted verification of new features
   - Debug UI issues in real-time
   - Verify keyboard shortcuts work correctly

3. **Playwright E2E Tests** - For regression testing
   - Convert verified features to automated tests
   - Run in CI/CD on every PR
   - Catch regressions before merge

### Test Data

For E2E tests, consider:

1. **Use existing dev database** - Tests run against real PocketBase data
2. **Seed test data** - Create fixtures before tests run
3. **Reset between tests** - Use API to clean up after each test

Example seed script (`ui/e2e/fixtures/seed.ts`):

```typescript
import PocketBase from 'pocketbase';

const pb = new PocketBase('http://localhost:8090');

export async function seedTestData() {
  // Create test board
  const board = await pb.collection('boards').create({
    name: 'Test Board',
    description: 'Board for E2E tests'
  });

  // Create test tasks
  await pb.collection('tasks').create({
    title: 'High Priority Bug',
    column: 'todo',
    priority: 'high',
    type: 'bug',
    board: board.id
  });

  await pb.collection('tasks').create({
    title: 'Low Priority Feature',
    column: 'backlog',
    priority: 'low',
    type: 'feature',
    board: board.id
  });

  return { boardId: board.id };
}

export async function cleanupTestData(boardId: string) {
  // Delete all tasks for this board
  const tasks = await pb.collection('tasks').getFullList({
    filter: `board = "${boardId}"`
  });
  for (const task of tasks) {
    await pb.collection('tasks').delete(task.id);
  }
  
  // Delete the board
  await pb.collection('boards').delete(boardId);
}
```

### Adding data-testid Attributes

For reliable E2E tests, add `data-testid` attributes to key elements:

```tsx
// SearchBar.tsx
<input
  data-testid="search-input"
  type="text"
  placeholder="Search..."
  // ...
/>

// FilterBuilder.tsx
<div data-testid="filter-builder" className={styles.modal}>
  <select data-testid="filter-field">
    {/* ... */}
  </select>
  <select data-testid="filter-operator">
    {/* ... */}
  </select>
  <button data-testid="add-filter-button">Add Filter</button>
</div>

// FilterBar.tsx
<button data-testid="clear-filters-button" onClick={clearFilters}>
  Clear all
</button>

// Board.tsx
<div data-testid="board-view">
  {/* ... */}
</div>

// ListView.tsx
<div data-testid="list-view">
  {/* ... */}
</div>
```

---

## Part 4: BrowserTools MCP (Alternative)

For reference, there's also BrowserTools MCP (`@agentdeskai/browser-tools-mcp`) which takes a different approach:

| Aspect | Playwright MCP | BrowserTools MCP |
|--------|---------------|------------------|
| **Purpose** | Control browser | Monitor browser |
| **Can click/type** | Yes | No |
| **Uses your browser** | No (new instance) | Yes (your session) |
| **Console/network logs** | Yes | Yes |
| **Lighthouse audits** | No | Yes |
| **Setup complexity** | Low (1 command) | High (extension + 2 servers) |
| **Maintainer** | Microsoft | Third-party |

BrowserTools is better for:
- Debugging issues in your logged-in browser session
- Running Lighthouse audits
- Monitoring console errors and network requests

Playwright MCP is better for:
- Automated testing flows
- Keyboard shortcut testing
- Full browser control

**For EgenSkriven testing, Playwright MCP is recommended** because we need to interact with the UI, not just observe it.

---

## Quick Reference

### Install Playwright E2E
```bash
cd ui
npm init playwright@latest
```

### Install Playwright MCP (OpenCode)
Add to `~/.config/opencode/opencode.json`:
```json
{
  "mcp": {
    "playwright": {
      "type": "local",
      "command": ["npx", "@playwright/mcp@latest"],
      "enabled": true
    }
  }
}
```

### Run E2E Tests
```bash
cd ui
npx playwright test           # Run all tests
npx playwright test --headed  # See browser
npx playwright test --ui      # Interactive UI
npx playwright show-report    # View report
```

### Key Documentation Links
- Playwright Docs: https://playwright.dev/docs/intro
- Playwright MCP: https://github.com/microsoft/playwright-mcp
- Writing Tests: https://playwright.dev/docs/writing-tests
- Web Server Config: https://playwright.dev/docs/test-webserver
- Locators: https://playwright.dev/docs/locators
- Assertions: https://playwright.dev/docs/test-assertions
