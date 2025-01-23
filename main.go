package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jung-kurt/gofpdf"
)

type Customer struct {
	Name     string
	Address  string
	Phone    string
	Adults   int
	Children int
}

type RentalItem struct {
	Description string
	Rate        float64
	Days        int
	FromDate    time.Time
	ToDate      time.Time
}

type Bill struct {
	BillNumber string
	Customer   Customer
	Items      []RentalItem
	Date       time.Time
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Daily Room Rental System")

	// Customer Details
	billNumberEntry := widget.NewEntry()
	billNumberEntry.SetPlaceHolder("Bill Number")

	customerNameEntry := widget.NewEntry()
	customerNameEntry.SetPlaceHolder("Customer Name")

	addressEntry := widget.NewMultiLineEntry()
	addressEntry.SetPlaceHolder("Customer Address")

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Phone Number")

	adultsEntry := widget.NewEntry()
	adultsEntry.SetPlaceHolder("Number of Adults")

	childrenEntry := widget.NewEntry()
	childrenEntry.SetPlaceHolder("Number of Children")

	// Room Details
	roomTypeSelect := widget.NewSelect([]string{
		"NON-AC Room",
		"AC Room",
	}, nil)

	rateEntry := widget.NewEntry()
	rateEntry.SetPlaceHolder("Rate per Day")

	fromDate := time.Now()
	toDate := time.Now()

	fromDatePicker := widget.NewEntry()
	fromDatePicker.SetText(fromDate.Format("02-01-2006"))
	fromDateButton := widget.NewButton("Select From Date", func() {
		// Create custom date selection
		year := widget.NewSelect(generateYears(), nil)
		month := widget.NewSelect(months(), nil)
		day := widget.NewSelect(generateDays(), nil)

		// Set current date as default
		year.SetSelected(fmt.Sprintf("%d", fromDate.Year()))
		month.SetSelected(fromDate.Month().String())
		day.SetSelected(fmt.Sprintf("%d", fromDate.Day()))

		content := container.NewVBox(
			widget.NewLabel("Select Date:"),
			container.NewGridWithColumns(3,
				container.NewVBox(
					widget.NewLabel("Day"),
					day,
				),
				container.NewVBox(
					widget.NewLabel("Month"),
					month,
				),
				container.NewVBox(
					widget.NewLabel("Year"),
					year,
				),
			),
		)

		d := dialog.NewCustom("Select From Date", "Set", content, myWindow)
		d.Resize(fyne.NewSize(300, 200))
		d.Show()

		// Update date when selections change
		year.OnChanged = func(value string) {
			updateFromDate(day.Selected, month.Selected, value, &fromDate, fromDatePicker)
		}
		month.OnChanged = func(value string) {
			updateFromDate(day.Selected, value, year.Selected, &fromDate, fromDatePicker)
		}
		day.OnChanged = func(value string) {
			updateFromDate(value, month.Selected, year.Selected, &fromDate, fromDatePicker)
		}
	})

	toDatePicker := widget.NewEntry()
	toDatePicker.SetText(toDate.Format("02-01-2006"))
	toDateButton := widget.NewButton("Select To Date", func() {
		// Create custom date selection
		year := widget.NewSelect(generateYears(), nil)
		month := widget.NewSelect(months(), nil)
		day := widget.NewSelect(generateDays(), nil)

		// Set current date as default
		year.SetSelected(fmt.Sprintf("%d", toDate.Year()))
		month.SetSelected(toDate.Month().String())
		day.SetSelected(fmt.Sprintf("%d", toDate.Day()))

		content := container.NewVBox(
			widget.NewLabel("Select Date:"),
			container.NewGridWithColumns(3,
				container.NewVBox(
					widget.NewLabel("Day"),
					day,
				),
				container.NewVBox(
					widget.NewLabel("Month"),
					month,
				),
				container.NewVBox(
					widget.NewLabel("Year"),
					year,
				),
			),
		)

		d := dialog.NewCustom("Select To Date", "Set", content, myWindow)
		d.Resize(fyne.NewSize(300, 200))
		d.Show()

		// Update date when selections change
		year.OnChanged = func(value string) {
			updateToDate(day.Selected, month.Selected, value, &toDate, toDatePicker)
		}
		month.OnChanged = func(value string) {
			updateToDate(day.Selected, value, year.Selected, &toDate, toDatePicker)
		}
		day.OnChanged = func(value string) {
			updateToDate(value, month.Selected, year.Selected, &toDate, toDatePicker)
		}
	})

	// Make the entry fields read-only
	fromDatePicker.Disable()
	toDatePicker.Disable()

	// Status label
	statusLabel := widget.NewLabel("")

	// Validate functions
	validateNumber := func(text string) error {
		_, err := strconv.Atoi(text)
		return err
	}

	validateDate := func(text string) time.Time {
		date, err := time.Parse("02-01-2006", text)
		if err != nil {
			return time.Time{}
		}
		return date
	}

	adultsEntry.Validator = validateNumber
	childrenEntry.Validator = validateNumber
	rateEntry.Validator = validateNumber

	// Store rental items
	var rentalItems []RentalItem

	// Display items
	itemsList := widget.NewTextGrid()
	updateItemsList := func() {
		text := "Rooms Booked:\n"
		for i, item := range rentalItems {
			text += fmt.Sprintf("%d. %s - ₹%.2f x %d days = ₹%.2f\n",
				i+1, item.Description, item.Rate, item.Days, item.Rate*float64(item.Days))
			text += fmt.Sprintf("   Period: %s to %s\n",
				item.FromDate.Format("02-01-2006"), item.ToDate.Format("02-01-2006"))
		}
		itemsList.SetText(text)
	}

	// Add Room Button
	addButton := widget.NewButton("Add Room", func() {
		rate, err := strconv.ParseFloat(rateEntry.Text, 64)
		if err != nil {
			statusLabel.SetText("Please enter a valid rate")
			return
		}

		fromDate := validateDate(fromDatePicker.Text)
		toDate := validateDate(toDatePicker.Text)
		if fromDate.IsZero() || toDate.IsZero() {
			statusLabel.SetText("Please enter valid dates (DD-MM-YYYY)")
			return
		}

		days := int(toDate.Sub(fromDate).Hours()/24) + 1
		if days < 1 {
			statusLabel.SetText("To Date must be after From Date")
			return
		}

		if roomTypeSelect.Selected == "" {
			statusLabel.SetText("Please select a room type")
			return
		}

		item := RentalItem{
			Description: roomTypeSelect.Selected,
			Rate:        rate,
			Days:        days,
			FromDate:    fromDate,
			ToDate:      toDate,
		}

		rentalItems = append(rentalItems, item)
		updateItemsList()

		// Clear room inputs
		roomTypeSelect.Selected = ""
		rateEntry.SetText("")
		fromDatePicker.SetText("")
		toDatePicker.SetText("")
		statusLabel.SetText("Room added successfully!")
	})

	// Generate Bill Button
	generateButton := widget.NewButton("Generate Bill", func() {
		if len(rentalItems) == 0 {
			statusLabel.SetText("Please add at least one room")
			return
		}

		adults, errA := strconv.Atoi(adultsEntry.Text)
		children, errC := strconv.Atoi(childrenEntry.Text)
		if errA != nil || errC != nil {
			statusLabel.SetText("Please enter valid numbers for adults and children")
			return
		}

		if billNumberEntry.Text == "" || customerNameEntry.Text == "" ||
			addressEntry.Text == "" || phoneEntry.Text == "" {
			statusLabel.SetText("Please fill in all customer details")
			return
		}

		customer := Customer{
			Name:     customerNameEntry.Text,
			Address:  addressEntry.Text,
			Phone:    phoneEntry.Text,
			Adults:   adults,
			Children: children,
		}

		bill := Bill{
			BillNumber: billNumberEntry.Text,
			Customer:   customer,
			Items:      rentalItems,
			Date:       time.Now(),
		}

		err := generatePDF(bill)
		if err != nil {
			statusLabel.SetText("Error generating PDF: " + err.Error())
			return
		}

		statusLabel.SetText("Bill generated successfully!")
	})

	// Clear All Button
	clearButton := widget.NewButton("Clear All", func() {
		billNumberEntry.SetText("")
		customerNameEntry.SetText("")
		addressEntry.SetText("")
		phoneEntry.SetText("")
		adultsEntry.SetText("")
		childrenEntry.SetText("")
		roomTypeSelect.Selected = ""
		rateEntry.SetText("")
		fromDatePicker.SetText("")
		toDatePicker.SetText("")
		rentalItems = nil
		updateItemsList()
		statusLabel.SetText("All fields cleared")
	})

	// Layout
	customerDetails := container.NewVBox(
		widget.NewLabel("Customer Details"),
		billNumberEntry,
		customerNameEntry,
		addressEntry,
		phoneEntry,
		adultsEntry,
		childrenEntry,
	)

	roomDetails := container.NewVBox(
		widget.NewLabel("Room Details"),
		roomTypeSelect,
		rateEntry,
		fromDateButton,
		fromDatePicker,
		toDateButton,
		toDatePicker,
		addButton,
	)

	buttons := container.NewHBox(
		generateButton,
		clearButton,
	)

	content := container.NewVBox(
		customerDetails,
		roomDetails,
		widget.NewLabel("\nBooked Rooms"),
		itemsList,
		buttons,
		statusLabel,
	)

	myWindow.SetContent(container.NewPadded(content))
	myWindow.Resize(fyne.NewSize(500, 800))
	myWindow.ShowAndRun()
}

func generatePDF(bill Bill) error {
	// Create Invoice directory if it doesn't exist
	if err := os.MkdirAll("Invoice", 0755); err != nil {
		return fmt.Errorf("failed to create Invoice directory: %v", err)
	}

	// Create filename with bill number
	filename := filepath.Join("Invoice", fmt.Sprintf("Invoice_%s.pdf", bill.BillNumber))

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Company Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "Trinity Stays")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, "123, Main Street, Chennai - 600001")
	pdf.Ln(5)
	pdf.Cell(190, 5, "Phone: +91 98765 43210")
	pdf.Ln(5)
	pdf.Cell(190, 5, "GSTIN: 33AALCT2345K1ZB")
	pdf.Ln(15)

	// Add line separator
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// Bill Details in a box
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 12)

	// Create a box for Bill Details
	startY := pdf.GetY()
	pdf.Rect(10, startY, 90, 8, "F")
	pdf.Cell(90, 8, "Bill Details")

	// Create a box for Customer Details
	pdf.Rect(105, startY, 90, 8, "F")
	pdf.Cell(5, 8, "") // spacing
	pdf.Cell(90, 8, "Customer Details")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	// Left side - Bill details with borders
	startY = pdf.GetY()
	pdf.Rect(10, startY, 90, 24, "D") // Border for bill details
	pdf.SetX(15)                      // Indent from left
	pdf.Cell(25, 6, "Bill No:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, bill.BillNumber)
	pdf.Ln(6)
	pdf.SetX(15)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, "Date:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, bill.Date.Format("02-01-2006"))
	pdf.Ln(6)

	// Right side - Customer details with borders
	pdf.Rect(105, startY, 90, 40, "D") // Border for customer details
	pdf.SetXY(110, startY)             // Indent from left of customer section
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, "Name:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, bill.Customer.Name)
	pdf.Ln(6)
	pdf.SetX(110)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, "Phone:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, bill.Customer.Phone)
	pdf.Ln(6)
	pdf.SetX(110)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, "Address:")
	pdf.SetFont("Arial", "B", 10)
	// currentY := pdf.GetY()
	pdf.MultiCell(60, 6, bill.Customer.Address, "", "", false)

	// Move to the maximum Y position used
	pdf.SetY(math.Max(pdf.GetY(), startY+45))
	pdf.Ln(5)

	// Guest Information with a light background
	pdf.SetFillColor(240, 240, 240)
	pdf.Rect(10, pdf.GetY(), 185, 8, "F")
	pdf.SetX(15)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 8, "No. of Guests:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(130, 8, fmt.Sprintf("%d Adults, %d Children",
		bill.Customer.Adults, bill.Customer.Children))
	pdf.Ln(12)

	// Add line separator
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// Table headers with filled background
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)

	// Create table header cells with border
	pdf.CellFormat(45, 8, "Room Type", "1", 0, "", true, 0, "")
	pdf.CellFormat(30, 8, "Rate/Day", "1", 0, "", true, 0, "")
	pdf.CellFormat(20, 8, "Days", "1", 0, "", true, 0, "")
	pdf.CellFormat(55, 8, "Period", "1", 0, "", true, 0, "")
	pdf.CellFormat(40, 8, "Amount", "1", 1, "", true, 0, "")

	// Items
	pdf.SetFont("Arial", "", 10)
	subtotal := 0.0
	for _, item := range bill.Items {
		amount := item.Rate * float64(item.Days)
		subtotal += amount

		period := fmt.Sprintf("%s to %s",
			item.FromDate.Format("02/01/06"), item.ToDate.Format("02/01/06"))

		pdf.CellFormat(45, 8, item.Description, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.2f", item.Rate), "1", 0, "", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%d", item.Days), "1", 0, "", false, 0, "")
		pdf.CellFormat(55, 8, period, "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", amount), "1", 1, "", false, 0, "")
	}

	// Totals section with right alignment
	pdf.Ln(5)
	gst := subtotal * 0.18
	total := subtotal + gst

	// Add line separator
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	// Right-aligned totals using CellFormat
	pdf.CellFormat(150, 8, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", subtotal), "", 1, "R", false, 0, "")

	pdf.CellFormat(150, 8, "GST (18%):", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", gst), "", 1, "R", false, 0, "")

	// Total amount with box
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(150, 8, "Total Amount:", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", total), "1", 1, "R", true, 0, "")
	pdf.Ln(15)

	// Terms and conditions
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Terms & Conditions:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(190, 4, "1. Check-in time is 12:00 PM and check-out time is 11:00 AM\n"+
		"2. Payment to be made in advance\n"+
		"3. No refunds for early check-out\n"+
		"4. ID proof is mandatory for all guests\n"+
		"5. The management is not responsible for any valuables", "", "", false)

	// Footer with signature
	pdf.Ln(10)
	pdf.Line(140, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(3)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(130, 4, "")
	pdf.Cell(60, 4, "Authorized Signature")

	return pdf.OutputFileAndClose(filename)
}

func generateYears() []string {
	currentYear := time.Now().Year()
	years := make([]string, 5)
	for i := 0; i < 5; i++ {
		years[i] = fmt.Sprintf("%d", currentYear+i)
	}
	return years
}

func generateDays() []string {
	days := make([]string, 31)
	for i := 1; i <= 31; i++ {
		days[i-1] = fmt.Sprintf("%d", i)
	}
	return days
}

func months() []string {
	return []string{
		"January", "February", "March", "April",
		"May", "June", "July", "August",
		"September", "October", "November", "December",
	}
}

func updateFromDate(day, month, year string, fromDate *time.Time, picker *widget.Entry) {
	monthNum := getMonthNumber(month)
	d, _ := strconv.Atoi(day)
	y, _ := strconv.Atoi(year)

	newDate := time.Date(y, time.Month(monthNum), d, 0, 0, 0, 0, time.Local)
	*fromDate = newDate
	picker.SetText(newDate.Format("02-01-2006"))
}

func updateToDate(day, month, year string, toDate *time.Time, picker *widget.Entry) {
	monthNum := getMonthNumber(month)
	d, _ := strconv.Atoi(day)
	y, _ := strconv.Atoi(year)

	newDate := time.Date(y, time.Month(monthNum), d, 0, 0, 0, 0, time.Local)
	*toDate = newDate
	picker.SetText(newDate.Format("02-01-2006"))
}

func getMonthNumber(month string) int {
	months := map[string]int{
		"January": 1, "February": 2, "March": 3, "April": 4,
		"May": 5, "June": 6, "July": 7, "August": 8,
		"September": 9, "October": 10, "November": 11, "December": 12,
	}
	return months[month]
}
