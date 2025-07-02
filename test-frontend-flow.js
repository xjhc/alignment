import { chromium } from 'playwright';

async function testFrontendFlow() {
  console.log('Starting frontend flow test...');
  
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    // Navigate to the frontend
    console.log('Navigating to frontend...');
    await page.goto('http://localhost:5173');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // Check if we get the navigation error
    const errorText = await page.textContent('body').catch(() => '');
    console.log('Page content includes navigation error?', errorText.includes('navigateToLobbyList'));
    
    // Take a screenshot for debugging
    await page.screenshot({ path: 'frontend-test.png' });
    console.log('Screenshot saved as frontend-test.png');
    
    // Check if login screen loads
    const loginButton = await page.locator('button:has-text("Login")').first();
    if (await loginButton.isVisible()) {
      console.log('✅ Login screen loaded successfully');
    } else {
      console.log('❌ Login screen not found');
    }
    
    // Try to interact with the login form
    const nameInput = await page.locator('input[placeholder*="name"], input[placeholder*="Name"]').first();
    if (await nameInput.isVisible()) {
      console.log('✅ Name input found');
      await nameInput.fill('TestPlayer');
    }
    
  } catch (error) {
    console.error('Test failed:', error);
  } finally {
    await browser.close();
  }
}

testFrontendFlow();