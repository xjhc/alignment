import { test, expect } from '@playwright/test';

test.describe('Frontend Navigation Fix', () => {
  test('should not show navigateToLobbyList initialization error', async ({ page }) => {
    // Navigate to the frontend
    await page.goto('http://localhost:5173');
    
    // Wait for initial load
    await page.waitForTimeout(2000);
    
    // Check that we don't get the navigation initialization error
    const bodyText = await page.textContent('body');
    expect(bodyText).not.toContain('Cannot access \'navigateToLobbyList\' before initialization');
    
    // Take a screenshot for debugging
    await page.screenshot({ path: 'navigation-fix-test.png' });
  });

  test('should display login screen without errors', async ({ page }) => {
    await page.goto('http://localhost:5173');
    
    // Wait for page to load and check for React error boundary
    await page.waitForTimeout(3000);
    
    // Should not show "Something went wrong" error
    const hasError = await page.locator('text=Something went wrong').isVisible();
    expect(hasError).toBe(false);
    
    // Should show login screen components
    const loginForm = await page.locator('[class*="launch-form"], [class*="login"]').first();
    await expect(loginForm).toBeVisible();
  });
});