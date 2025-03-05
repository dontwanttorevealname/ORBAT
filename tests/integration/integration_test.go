package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
	"orbat/internal/database"
)

var pw *playwright.Playwright
var browser playwright.Browser
var isHeadless bool

func init() {
	// Check if we're running in CI/CD environment
	if os.Getenv("CI") != "" {
		isHeadless = true
	} else {
		// Default to false for local development
		isHeadless = false
	}
}

func ensurePlaywrightInstalled() error {
	// Check if playwright CLI is installed
	_, err := exec.LookPath("playwright")
	if err != nil {
		// Install playwright CLI
		cmd := exec.Command("go", "install", "github.com/playwright-community/playwright-go/cmd/playwright@latest")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install playwright CLI: %v", err)
		}
	}

	// Install browser binaries
	cmd := exec.Command("playwright", "install")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install browser binaries: %v", err)
	}

	return nil
}

func TestMain(m *testing.M) {
	// Ensure Playwright is installed
	if err := ensurePlaywrightInstalled(); err != nil {
		panic(fmt.Sprintf("Failed to setup Playwright: %v", err))
	}

	// Load test environment variables
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic("Error loading .env.test file")
	}

	// Initialize database connection
	if err := database.Initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	// Start Playwright
	pwt, err := playwright.Run()
	if err != nil {
		panic(fmt.Sprintf("Failed to start Playwright: %v", err))
	}
	pw = pwt

	// Launch browser with configurable headless mode
	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(isHeadless),
		SlowMo:   playwright.Float(100),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to launch browser: %v", err))
	}

	// Create a new browser context with a larger viewport
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Screen: &playwright.Size{
			Width:  1280,
			Height: 720,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create context: %v", err))
	}

	// Use the context to create pages
	page, err := context.NewPage()
	if err != nil {
		panic(fmt.Sprintf("Failed to create page: %v", err))
	}
	defer page.Close()

	// Run tests
	code := m.Run()

	// Cleanup
	if err := browser.Close(); err != nil {
		fmt.Printf("Failed to close browser: %v\n", err)
	}
	if err := pw.Stop(); err != nil {
		fmt.Printf("Failed to stop Playwright: %v\n", err)
	}

	os.Exit(code)
}

// Helper function to create a new page for each test
func newPage(t *testing.T) playwright.Page {
	// Create a new context for each test to ensure clean state
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Screen: &playwright.Size{
			Width:  1280,
			Height: 720,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}

	// Add a cleanup function to close the context when the test finishes
	t.Cleanup(func() {
		if err := context.Close(); err != nil {
			t.Errorf("Failed to close context: %v", err)
		}
	})

	return page
}

func TestGroupsList(t *testing.T) {
	t.Log("Starting TestGroupsList")
	page := newPage(t)
	defer page.Close()

	t.Log("Navigating to groups page")
	if _, err := page.Goto("http://localhost:8080"); err != nil {
		t.Fatalf("Failed to navigate to groups page: %v", err)
	}
	t.Log("Successfully loaded groups page")

	t.Log("Attempting to highlight Ranger Rifle Squad element")
	if _, err := page.Evaluate(`() => {
		const el = document.evaluate("//text()[contains(., 'Ranger Rifle Squad')]", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue.parentElement;
		el.style.border = '2px solid red';
		el.style.backgroundColor = 'yellow';
	}`); err != nil {
		t.Logf("Failed to highlight element: %v", err)
	}
	t.Log("Successfully highlighted element")

	// Check for Ranger Rifle Squad (from seed data)
	text, err := page.TextContent("text='Ranger Rifle Squad'")
	if err != nil {
		t.Fatalf("Failed to find Ranger Rifle Squad: %v", err)
	}
	if text == "" {
		t.Error("Ranger Rifle Squad not found on page")
	}

	// Click on a group name link (using the actual structure from groups.html)
	groupLink, err := page.QuerySelector(".card-title a")
	if err != nil {
		t.Fatalf("Failed to find group link: %v", err)
	}
	
	groupName, err := groupLink.TextContent()
	if err != nil {
		t.Fatalf("Failed to get group name: %v", err)
	}

	if err := groupLink.Click(); err != nil {
		t.Fatalf("Failed to click group link: %v", err)
	}

	// Verify we're on the group details page
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for page load: %v", err)
	}

	titleElement, err := page.QuerySelector("h1")
	if err != nil {
		t.Fatalf("Failed to find title element: %v", err)
	}

	title, err := titleElement.TextContent()
	if err != nil {
		t.Fatalf("Failed to get title text: %v", err)
	}

	if !strings.Contains(title, groupName) {
		t.Errorf("Expected group details page title to contain %q, got %q", groupName, title)
	}

	// Click back button
	backButton, err := page.QuerySelector("a.btn-outline-primary")
	if err != nil {
		t.Fatalf("Failed to find back button: %v", err)
	}

	if err := backButton.Click(); err != nil {
		t.Fatalf("Failed to click back button: %v", err)
	}

	// Verify we're back on groups list page
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for page load: %v", err)
	}

	titleElement, err = page.QuerySelector("h1")
	if err != nil {
		t.Fatalf("Failed to find title element: %v", err)
	}

	title, err = titleElement.TextContent()
	if err != nil {
		t.Fatalf("Failed to get title text: %v", err)
	}

	if title != "Military Groups" {
		t.Errorf("Expected to be back on groups list page, got title %q", title)
	}
}

func TestWeaponsList(t *testing.T) {
	t.Log("Starting TestWeaponsList")
	page := newPage(t)

	t.Log("Navigating to weapons page")
	if _, err := page.Goto("http://localhost:8080/weapons"); err != nil {
		t.Fatalf("Failed to navigate to weapons page: %v", err)
	}
	t.Log("Successfully loaded weapons page")

	t.Log("Checking for existing M4A1 weapon")
	text, err := page.TextContent("text='M4A1'")
	if err != nil {
		t.Fatalf("Failed to find M4A1: %v", err)
	}
	if text == "" {
		t.Error("M4A1 not found on page")
	}
	t.Log("Successfully found M4A1 on page")

	// Check caliber
	text, err = page.TextContent("text='5.56mm'")
	if err != nil {
		t.Fatalf("Failed to find caliber: %v", err)
	}
	if text == "" {
		t.Error("5.56mm caliber not found on page")
	}

	// Create a new weapon
	weaponName := "Test M16A4"
	weaponType := "Assault Rifle"
	weaponCaliber := "5.56x45mm NATO"

	// Fill out the form
	if err := page.Fill("#name", weaponName); err != nil {
		t.Fatalf("Failed to fill weapon name: %v", err)
	}
	if err := page.Fill("#type", weaponType); err != nil {
		t.Fatalf("Failed to fill weapon type: %v", err)
	}
	if err := page.Fill("#caliber", weaponCaliber); err != nil {
		t.Fatalf("Failed to fill weapon caliber: %v", err)
	}

	// Submit the form
	submitButton, err := page.QuerySelector("button[type='submit']")
	if err != nil {
		t.Fatalf("Failed to find submit button: %v", err)
	}

	if err := submitButton.Click(); err != nil {
		t.Fatalf("Failed to click submit button: %v", err)
	}

	// Wait for navigation after form submission
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after form submission: %v", err)
	}

	// Verify the weapon appears in the list
	text, err = page.TextContent(fmt.Sprintf("text='%s'", weaponName))
	if err != nil {
		t.Fatalf("Failed to find new weapon: %v", err)
	}
	if text == "" {
		t.Error("New weapon not found on page")
	}

	// Click on the weapon details link
	// First find all cards
	cards, err := page.QuerySelectorAll(".card")
	if err != nil {
		t.Fatalf("Failed to find weapon cards: %v", err)
	}

	var detailsLink playwright.ElementHandle
	// Look through each card for our weapon
	for _, card := range cards {
		cardText, err := card.TextContent()
		if err != nil {
			continue
		}
		if strings.Contains(cardText, weaponName) {
			// Found our card, now find the details link
			detailsLink, err = card.QuerySelector("a.btn-outline-primary")
			if err != nil {
				t.Fatalf("Failed to find details link in card: %v", err)
			}
			break
		}
	}

	if detailsLink == nil {
		t.Fatalf("Could not find details link for weapon %s", weaponName)
	}

	if err := detailsLink.Click(); err != nil {
		t.Fatalf("Failed to click weapon details link: %v", err)
	}

	// Wait for navigation to details page
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation to details page: %v", err)
	}

	titleElement, err := page.QuerySelector("h1")
	if err != nil {
		t.Fatalf("Failed to find title element: %v", err)
	}

	title, err := titleElement.TextContent()
	if err != nil {
		t.Fatalf("Failed to get title text: %v", err)
	}

	if !strings.Contains(title, weaponName) {
		t.Errorf("Expected weapon details page title to contain %q, got %q", weaponName, title)
	}

	// Delete the weapon
	dialogCount := 0
	page.On("dialog", func(dialog playwright.Dialog) {
		dialogCount++
		t.Logf("Handling dialog %d: %s", dialogCount, dialog.Message())
		
		if dialogCount == 1 {
			// First dialog is the confirmation
			t.Log("Accepting confirmation dialog")
			dialog.Accept()
		} else {
			// Second dialog is the password prompt
			t.Log("Entering password in dialog")
			dialog.Accept("adminpassword")
		}
	})

	// Find and click delete button
	deleteButton, err := page.QuerySelector("button.btn-danger")
	if err != nil {
		t.Fatalf("Failed to find delete button: %v", err)
	}

	if err := deleteButton.Click(); err != nil {
		t.Fatalf("Failed to click delete button: %v", err)
	}

	// Wait for navigation after deletion
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after deletion: %v", err)
	}

	// Final verification that the weapon was deleted
	text, err = page.TextContent(fmt.Sprintf("text='%s'", weaponName))
	if err == nil && text != "" {
		t.Error("Weapon still found on page after deletion")
	}

	// Explicitly close the page and its context
	if err := page.Close(); err != nil {
		t.Errorf("Failed to close page: %v", err)
	}
	
	context := page.Context()
	if err := context.Close(); err != nil {
		t.Errorf("Failed to close context: %v", err)
	}

	t.Log("Test completed successfully")
}

func TestCountriesList(t *testing.T) {
	t.Log("Starting TestCountriesList")
	page := newPage(t)
	defer page.Close()

	t.Log("Navigating to countries page")
	if _, err := page.Goto("http://localhost:8080/countries"); err != nil {
		t.Fatalf("Failed to navigate to countries page: %v", err)
	}
	t.Log("Successfully loaded countries page")

	// Highlight and check for USA
	if _, err := page.Evaluate(`() => {
		const el = document.evaluate("//text()[contains(., 'United States of America')]", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue.parentElement;
		el.style.border = '2px solid red';
		el.style.backgroundColor = 'yellow';
	}`); err != nil {
		t.Logf("Failed to highlight element: %v", err)
	}

	text, err := page.TextContent("text='United States of America'")
	if err != nil {
		t.Fatalf("Failed to find USA: %v", err)
	}
	if text == "" {
		t.Error("United States of America not found on page")
	}

	// Check for United Kingdom
	text, err = page.TextContent("text='United Kingdom'")
	if err != nil {
		t.Fatalf("Failed to find UK: %v", err)
	}
	if text == "" {
		t.Error("United Kingdom not found on page")
	}
}

func TestCountryDetails(t *testing.T) {
	page := newPage(t)
	defer page.Close()

	// Navigate to USA details
	if _, err := page.Goto("http://localhost:8080/country/United%20States%20of%20America"); err != nil {
		t.Fatalf("Failed to navigate to country details: %v", err)
	}

	// Highlight and check for Ranger Rifle Squad
	if _, err := page.Evaluate(`() => {
		const el = document.evaluate("//text()[contains(., 'Ranger Rifle Squad')]", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue.parentElement;
		el.style.border = '2px solid red';
		el.style.backgroundColor = 'yellow';
	}`); err != nil {
		t.Logf("Failed to highlight element: %v", err)
	}

	text, err := page.TextContent("text='Ranger Rifle Squad'")
	if err != nil {
		t.Fatalf("Failed to find Ranger Rifle Squad: %v", err)
	}
	if text == "" {
		t.Error("Ranger Rifle Squad not found on country page")
	}

	// Check for weapon usage
	text, err = page.TextContent("text='M4A1'")
	if err != nil {
		t.Fatalf("Failed to find M4A1 usage: %v", err)
	}
	if text == "" {
		t.Error("M4A1 not found on country page")
	}
}

func TestGroupCreationAndDeletion(t *testing.T) {
	t.Log("Starting TestGroupCreationAndDeletion")
	page := newPage(t)

	t.Log("Navigating to groups page")
	if _, err := page.Goto("http://localhost:8080"); err != nil {
		t.Fatalf("Failed to navigate to groups page: %v", err)
	}
	t.Log("Successfully loaded groups page")

	t.Log("Waiting for page to be ready")
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for page load: %v", err)
	}
	t.Log("Page is ready")

	// Click "Add New Group" button
	addButton, err := page.QuerySelector("a:text('Add New Group')")
	if err != nil {
		t.Fatalf("Failed to find Add New Group button: %v", err)
	}
	if addButton == nil {
		content, _ := page.Content()
		t.Logf("Page content: %s", content)
		t.Fatal("Add New Group button not found")
	}
	if err := addButton.Click(); err != nil {
		t.Fatalf("Failed to click Add New Group button: %v", err)
	}

	// Wait for navigation to the add group page
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation to add page: %v", err)
	}

	// Create a new group
	groupName := "Test Infantry Squad"
	countryName := "United States of America"

	// Fill out the form
	nameInput, err := page.QuerySelector("#name")
	if err != nil {
		t.Fatalf("Failed to find name input: %v", err)
	}
	if nameInput == nil {
		content, _ := page.Content()
		t.Logf("Add page content: %s", content)
		t.Fatal("Name input element not found")
	}
	if err := nameInput.Fill(groupName); err != nil {
		t.Fatalf("Failed to fill group name: %v", err)
	}
	t.Log("Filled name input")

	// Fill nationality
	nationalityInput, err := page.QuerySelector("#nationality")
	if err != nil {
		t.Fatalf("Failed to find nationality input: %v", err)
	}
	if nationalityInput == nil {
		t.Fatal("Nationality input element not found")
	}
	if err := nationalityInput.Fill(countryName); err != nil {
		t.Fatalf("Failed to fill nationality: %v", err)
	}
	t.Log("Filled nationality")

	// Add a member (the form requires at least one)
	// The first member is added automatically by the page's JavaScript

	// Fill out the first member's details
	if err := page.Fill("input[name='role[]']", "Squad Leader"); err != nil {
		t.Fatalf("Failed to fill member role: %v", err)
	}
	if err := page.Fill("input[name='rank[]']", "Sergeant"); err != nil {
		t.Fatalf("Failed to fill member rank: %v", err)
	}

	// Select a weapon for the member (using the first weapon option)
	if _, err := page.SelectOption("select[name='weapons_0[]']", playwright.SelectOptionValues{
		Values: &[]string{"1"}, // M4A1
	}); err != nil {
		t.Fatalf("Failed to select weapon: %v", err)
	}

	// Submit the form
	submitButton, err := page.QuerySelector("button[type='submit']")
	if err != nil {
		t.Fatalf("Failed to find submit button: %v", err)
	}
	if submitButton == nil {
		t.Fatal("Submit button not found")
	}
	if err := submitButton.Click(); err != nil {
		t.Fatalf("Failed to submit form: %v", err)
	}
	t.Log("Submitted form")

	// Wait for navigation
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after form submission: %v", err)
	}

	// Verify the group appears in the list
	groupCard, err := page.QuerySelector(fmt.Sprintf(".card:has-text('%s')", groupName))
	if err != nil {
		t.Fatalf("Failed to find new group card: %v", err)
	}
	if groupCard == nil {
		t.Fatal("New group card not found on page")
	}

	// Click on the group details link
	detailsLink, err := groupCard.QuerySelector("a.btn-outline-primary")
	if err != nil {
		t.Fatalf("Failed to find details link: %v", err)
	}
	if err := detailsLink.Click(); err != nil {
		t.Fatalf("Failed to click details link: %v", err)
	}

	// Wait for navigation
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation to details page: %v", err)
	}

	// Delete the group
	// Set up dialog handlers
	dialogCount := 0
	page.On("dialog", func(dialog playwright.Dialog) {
		dialogCount++
		t.Logf("Handling dialog %d: %s", dialogCount, dialog.Message())
		
		if dialogCount == 1 {
			// First dialog is the confirmation
			t.Log("Accepting confirmation dialog")
			dialog.Accept()
		} else {
			// Second dialog is the password prompt
			t.Log("Entering password in dialog")
			dialog.Accept("adminpassword")
		}
	})
	
	// Find and click delete button
	deleteButton, err := page.QuerySelector("button.btn-danger")
	if err != nil {
		t.Fatalf("Failed to find delete button: %v", err)
	}
	if err := deleteButton.Click(); err != nil {
		t.Fatalf("Failed to click delete button: %v", err)
	}

	// Wait for navigation after deletion
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after deletion: %v", err)
	}

	// Verify the group is no longer in the list
	groupCard, err = page.QuerySelector(fmt.Sprintf(".card:has-text('%s')", groupName))
	if err != nil {
		t.Fatalf("Failed to query for group card: %v", err)
	}
	if groupCard != nil {
		t.Error("Group still found on page after deletion")
	}

	// Cleanup
	if err := page.Close(); err != nil {
		t.Errorf("Failed to close page: %v", err)
	}
	
	context := page.Context()
	if err := context.Close(); err != nil {
		t.Errorf("Failed to close context: %v", err)
	}

	t.Log("Test completed successfully")
}

func TestVehicleCreationAndDeletion(t *testing.T) {
	t.Log("Starting TestVehicleCreationAndDeletion")
	page := newPage(t)
	defer page.Close()

	t.Log("Navigating to vehicles page")
	if _, err := page.Goto("http://localhost:8080/vehicles"); err != nil {
		t.Fatalf("Failed to navigate to vehicles page: %v", err)
	}
	t.Log("Successfully loaded vehicles page")

	t.Log("Waiting for page to be ready")
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for page load: %v", err)
	}
	t.Log("Page is ready")

	// Create a new vehicle
	vehicleName := "Test Bradley IFV"
	vehicleType := "Infantry Fighting Vehicle"
	vehicleArmament := "25mm M242 Bushmaster Chain Gun, TOW ATGM"

	t.Log("Filling out vehicle creation form")
	if err := page.Fill("#name", vehicleName); err != nil {
		t.Fatalf("Failed to fill vehicle name: %v", err)
	}
	t.Log("Successfully filled vehicle name")

	if err := page.Fill("#type", vehicleType); err != nil {
		t.Fatalf("Failed to fill vehicle type: %v", err)
	}
	t.Log("Successfully filled vehicle type")

	if err := page.Fill("#armament", vehicleArmament); err != nil {
		t.Fatalf("Failed to fill vehicle armament: %v", err)
	}
	t.Log("Successfully filled vehicle armament")

	t.Log("Setting up dialog handler")
	page.On("dialog", func(dialog playwright.Dialog) {
		t.Logf("Handling dialog: %s", dialog.Message())
		dialog.Accept()
	})
	t.Log("Dialog handler set up successfully")

	t.Log("Looking for submit button")
	submitButton, err := page.QuerySelector("#vehicleForm button[type='submit']")
	if err != nil {
		t.Fatalf("Failed to find submit button: %v", err)
	}
	if submitButton == nil {
		t.Fatal("Submit button not found")
	}
	t.Log("Found submit button")

	t.Log("Submitting vehicle creation form")
	if err := submitButton.Click(); err != nil {
		t.Fatalf("Failed to click submit button: %v", err)
	}
	t.Log("Form submitted successfully")

	t.Log("Waiting for navigation after form submission")
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after form submission: %v", err)
	}
	t.Log("Navigation completed")

	t.Log("Verifying new vehicle appears in list")
	selector := fmt.Sprintf(".card-title:has-text('%s')", vehicleName)
	if _, err := page.WaitForSelector(selector); err != nil {
		content, _ := page.Content()
		t.Logf("Page content after submission: %s", content)
		t.Fatalf("Failed to find new vehicle card title: %v", err)
	}
	t.Log("Successfully found new vehicle in list")

	t.Log("Looking for details link")
	detailsLink, err := page.QuerySelector(fmt.Sprintf(".card:has-text('%s') a.btn-outline-primary", vehicleName))
	if err != nil {
		t.Fatalf("Failed to find details link: %v", err)
	}
	if detailsLink == nil {
		content, _ := page.Content()
		t.Logf("Page content when looking for details link: %s", content)
		t.Fatal("Details link not found")
	}
	t.Log("Found details link")

	t.Log("Clicking details link")
	if err := detailsLink.Click(); err != nil {
		t.Fatalf("Failed to click details link: %v", err)
	}
	t.Log("Successfully clicked details link")

	t.Log("Waiting for navigation to details page")
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation to details page: %v", err)
	}
	t.Log("Successfully navigated to details page")

	t.Log("Verifying vehicle details")
	titleText, err := page.TextContent("h1")
	if err != nil {
		t.Fatalf("Failed to find title: %v", err)
	}
	if !strings.Contains(titleText, vehicleName) {
		t.Errorf("Expected title to contain %q, got %q", vehicleName, titleText)
	}
	t.Log("Successfully verified vehicle details")

	t.Log("Setting up dialog handlers for deletion")
	page.On("dialog", func(dialog playwright.Dialog) {
		t.Logf("Handling dialog: %s", dialog.Message())
		if strings.Contains(dialog.Message(), "Are you sure") {
			t.Log("Accepting deletion confirmation dialog")
			dialog.Accept()
		} else if strings.Contains(dialog.Message(), "password") {
			t.Log("Entering password in dialog")
			dialog.Accept("adminpassword")
		}
	})
	t.Log("Dialog handlers set up successfully")

	t.Log("Looking for delete button")
	deleteButton, err := page.QuerySelector("form[method='POST'][action*='/delete'] button.btn-danger")
	if err != nil {
		t.Fatalf("Failed to find delete button: %v", err)
	}
	if deleteButton == nil {
		content, _ := page.Content()
		t.Logf("Page content when looking for delete button: %s", content)
		t.Fatal("Delete button not found")
	}
	t.Log("Found delete button")

	t.Log("Clicking delete button")
	if err := deleteButton.Click(); err != nil {
		t.Fatalf("Failed to click delete button: %v", err)
	}
	t.Log("Successfully clicked delete button")

	t.Log("Waiting for navigation after deletion")
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("Failed to wait for navigation after deletion: %v", err)
	}
	t.Log("Navigation completed")

	t.Log("Verifying vehicle was deleted")
	if _, err := page.WaitForSelector(fmt.Sprintf(".card-title:has-text('%s')", vehicleName), playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateDetached,
	}); err != nil {
		t.Error("Vehicle still found on page after deletion")
	}
	t.Log("Successfully verified vehicle was deleted")

	t.Log("Test completed successfully")
} 