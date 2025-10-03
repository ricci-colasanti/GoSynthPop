package main

import (
	"bytes"
	"image"
	"image/color"

	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	fyneCanvas "fyne.io/fyne/v2/canvas" // Rename to avoid conflict
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// createScatterPlot creates a scatter plot using gonum and returns it as Fyne image
func createScatterPlot() *fyneCanvas.Image {
	// Create a new plot
	p := plot.New()
	p.Title.Text = "Scatter Plot with Gonum"
	p.X.Label.Text = "X Values"
	p.Y.Label.Text = "Y Values"

	// Create random scatter data
	rand.Seed(time.Now().UnixNano())
	pts := make(plotter.XYs, 50)
	for i := range pts {
		pts[i].X = rand.Float64() * 100
		pts[i].Y = rand.Float64() * 100
	}

	// Create scatter plotter
	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Color = color.RGBA{R: 255, A: 255} // Red points
	scatter.GlyphStyle.Radius = vg.Points(3)

	// Add scatter to plot
	p.Add(scatter)

	// Add a line of best fit (optional)
	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	line.LineStyle.Width = vg.Points(1)
	line.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	line.LineStyle.Color = color.RGBA{B: 255, A: 255} // Blue dashed line
	p.Add(line)

	// Render plot to image
	img := renderPlotToImage(p)
	return img
}

// renderPlotToImage converts a gonum plot to a Fyne image
func renderPlotToImage(p *plot.Plot) *fyneCanvas.Image {
	// Create a writer to capture PNG data
	var buf bytes.Buffer

	// Render plot to PNG
	width := vg.Inch * 6
	height := vg.Inch * 4
	canvas, err := p.WriterTo(width, height, "png")
	if err != nil {
		panic(err)
	}

	_, err = canvas.WriteTo(&buf)
	if err != nil {
		panic(err)
	}

	// Convert PNG bytes to image
	imgData, _, err := image.Decode(&buf)
	if err != nil {
		panic(err)
	}

	// Create Fyne image with renamed import
	fyneImg := fyneCanvas.NewImageFromImage(imgData)
	fyneImg.SetMinSize(fyne.NewSize(400, 300))
	fyneImg.FillMode = fyneCanvas.ImageFillContain

	return fyneImg
}

// createCustomScatter with specific data
func createCustomScatter(xData, yData []float64, title string) *fyneCanvas.Image {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "X Axis"
	p.Y.Label.Text = "Y Axis"

	// Create XYs from data
	pts := make(plotter.XYs, len(xData))
	for i := range xData {
		pts[i].X = xData[i]
		pts[i].Y = yData[i]
	}

	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Color = color.RGBA{G: 100, B: 200, A: 255}
	scatter.GlyphStyle.Radius = vg.Points(4)

	p.Add(scatter)

	// Add grid
	p.Add(plotter.NewGrid())

	return renderPlotToImage(p)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Gonum + Fyne Scatter Plots")
	myWindow.Resize(fyne.NewSize(600, 500))

	// Create initial scatter plot
	scatterImg := createScatterPlot()

	// Create sample data for custom plot
	xData := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	yData := []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	customImg := createCustomScatter(xData, yData, "Linear Relationship")

	// Refresh button
	refreshBtn := widget.NewButton("Generate New Random Data", func() {
		newImg := createScatterPlot()
		scatterImg.Image = newImg.Image
		scatterImg.Refresh()
	})

	// Tab container for multiple plots
	tabs := container.NewAppTabs(
		container.NewTabItem("Random Data", container.NewVBox(
			scatterImg,
			refreshBtn,
		)),
		container.NewTabItem("Linear Data", customImg),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
