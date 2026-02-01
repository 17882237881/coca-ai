import { test, expect } from '@playwright/test';

// Use specific user for testing
const email = `test+${Date.now()}@example.com`;
const password = 'password123';

test('User can signup, login, and logout', async ({ page }) => {
    // Handle alerts automatically
    page.on('dialog', async dialog => {
        console.log(`Dialog message: ${dialog.message()}`);
        await dialog.accept();
    });

    // 1. Signup
    await page.goto('http://localhost:5173/signup');
    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    // Confirm Password
    const passwordInputs = await page.locator('input[type="password"]').all();
    await passwordInputs[1].fill(password);

    await page.click('button[type="submit"]');

    // Expect redirect to login or alert (handling js alert might be tricky, let's assume redirect or prompt)
    // Our code does `router.push('/login')` on success.
    await expect(page).toHaveURL('http://localhost:5173/login');

    // 2. Login
    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    await page.click('button[type="submit"]');

    // Expect redirect to Chat
    await expect(page).toHaveURL('http://localhost:5173/chat');
    await expect(page.locator('text=Coca AI')).toBeVisible();

    // 3. Logout
    await page.click('text=Logout');

    // Expect redirect to Login
    await expect(page).toHaveURL('http://localhost:5173/login');
});
