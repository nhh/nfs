package tui

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app            *tview.Application
	leftSidebar    *tview.Flex
	rightSidebar   *tview.Flex
	podList        *tview.TextView
	logList        *tview.TextView
	commandView    *tview.TextView
	globalSettings *tview.TextView
	syncerView     *tview.TreeView

	onUpdateCh = make(chan string)
	onErrorCh  = make(chan string)
)

func GetUpdateChannel() chan<- string {
	return onUpdateCh
}

func GetErrorChannel() chan<- string {
	return onErrorCh
}

func DisplayApp() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func init() {
	app = tview.NewApplication()

	columns := tview.NewFlex().SetDirection(tview.FlexRow)

	globalSettings = scrollableGlobalSettingsView()
	syncerView = scrollableSyncerView()

	leftSidebar = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(globalSettings, 0, 1, true).
		AddItem(syncerView, 0, 1, false)

	podList = scrollableTextView()
	logList = scrollableLogView()

	rightSidebar = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(podList, 3, 1, false).
		AddItem(logList, 0, 1, false)

	flex := tview.NewFlex().
		AddItem(leftSidebar, 0, 3, false).
		AddItem(rightSidebar, 0, 5, false)

	commandView = scrollableCommandView()

	columns.
		AddItem(flex, 0, 1, false).
		AddItem(commandView, 3, 1, false)

	// Key-Handler: Tab zum Wechseln
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(podList)
			podList.SetBorderColor(tcell.ColorGreen)
			return nil
		case tcell.KeyEscape: // Beenden
			app.Stop()
			return nil
		}
		return event
	})

	app.SetRoot(columns, true)

	go updateTitle()
}

func updateTitle() {
	for {
		select {
		case msg1 := <-onUpdateCh:
			app.QueueUpdateDraw(func() {
				logs := bytes.NewBufferString(logList.GetText(false))
				logs.WriteString(msg1)
				logList.SetText(logs.String())
				logList.ScrollToEnd()
			})
		case msg2 := <-onErrorCh:
			app.QueueUpdateDraw(func() {
				logs := bytes.NewBufferString(logList.GetText(false))
				logs.WriteString(msg2)
				logList.SetText(logs.String())
				logList.ScrollToEnd()
			})
		}
	}
}

func scrollableGlobalSettingsView() *tview.TextView {

	view := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)

	view.SetBorder(true).SetTitle("Config")

	yml := `
manifest: "v1"
# Events are grouped every <interval> for deduplication purposes
interval: 1000
# Pod specific configuration
pod:
  namespace: "fe-nihanft"
  selector: "app.kubernetes.io/name=frontend"
  cwd: "/home/frontend/"
watch:
  - pattern: "./**/*.php"
    excludes:
      - "node_modules"
    hooks:
      - "yarn run build"
  - pattern: "*.go"
`

	out, err := glamour.Render(fmt.Sprintf("```yml\n%s```", yml), "dark")

	if err != nil {
		panic(err)
	}

	view.SetText(tview.TranslateANSI(out))

	return view
}

func scrollableSyncerView() *tview.TreeView {
	// Wurzel-Knoten
	root := tview.NewTreeNode("-").
		SetColor(tview.Styles.TitleColor) // Farbe für den Root-Node

	// Erster Haupt-Knoten mit zwei Kindern
	node1 := tview.NewTreeNode("**/*.php").
		SetColor(tview.Styles.PrimaryTextColor)
	child1_1 := tview.NewTreeNode("app/Console/cake cache clear > /dev/null 2>&1").
		SetColor(tview.Styles.SecondaryTextColor)
	node1.AddChild(child1_1)

	// Zweiter Haupt-Knoten mit zwei Kindern
	node2 := tview.NewTreeNode("**/*.vue").
		SetColor(tview.Styles.PrimaryTextColor)
	child2_1 := tview.NewTreeNode("yarn run build").
		SetColor(tview.Styles.SecondaryTextColor)
	node2.AddChild(child2_1)

	// Zweiter Haupt-Knoten mit zwei Kindern
	node3 := tview.NewTreeNode("**/*.{css,html,woff}").
		SetColor(tview.Styles.PrimaryTextColor)
	child3_1 := tview.NewTreeNode("yarn run build").
		SetColor(tview.Styles.SecondaryTextColor)
	node3.AddChild(child3_1)

	// Vierter Haupt-Knoten mit zwei Kindern
	node4 := tview.NewTreeNode("lang/*.po").
		SetColor(tview.Styles.PrimaryTextColor)
	child4_1 := tview.NewTreeNode("yarn run build").
		SetColor(tview.Styles.SecondaryTextColor)
	node4.AddChild(child4_1)

	// Füge die Haupt-Knoten zur Wurzel hinzu
	root.AddChild(node1).AddChild(node2).AddChild(node3).AddChild(node4)

	// Erstelle ein TreeView mit der Wurzel
	treeView := tview.NewTreeView().
		SetRoot(root).       // Setze den Root-Node
		SetCurrentNode(root) // Standardmäßig fokussierter Knoten

	// Layout mit einer Border
	treeView.SetBorder(true).SetTitle("TreeView Example")

	// Ereignis-Handler für Auswahl
	treeView.SetSelectedFunc(func(node *tview.TreeNode) {
		nodeText := node.GetText()
		node.SetExpanded(!node.IsExpanded()) // Knoten ein-/ausklappen
		treeView.SetTitle("Selected: " + nodeText)
	})

	return treeView
}

func scrollableCommandView() *tview.TextView {
	// Daten für die horizontale Liste
	items := []string{
		"<a> Activate",
		"<b> Activate",
		"<c> Activate",
		"<d> Activate",
	}

	// Kombiniere die Items in einer Zeile
	listContent := ""
	for _, item := range items {
		listContent += item + "   " // Trennzeichen für Sichtbarkeit
	}

	// Erstelle ein neues TextView
	textView := tview.NewTextView().
		SetText(listContent).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	textView.SetBorder(true).SetTitle("Commands").SetTitleAlign(tview.AlignLeft)

	// Initiale Scroll-Position
	scrollX := 0

	// Ereignis-Handling für das Scrollen
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyRight: // Nach rechts scrollen
			scrollX++
		case tcell.KeyLeft: // Nach links scrollen
			if scrollX > 0 {
				scrollX--
			}
		case tcell.KeyEscape: // Beenden
			return nil
		default:
			break
		}

		textView.ScrollTo(scrollX, 0)

		return event
	})

	return textView
}

func scrollableLogView() *tview.TextView {
	// Erstelle ein neues TextView
	textView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	textView.SetBorder(true).SetTitle("Logs").SetTitleAlign(tview.AlignLeft)

	// Initiale Scroll-Position
	scrollY := 0

	// Ereignis-Handling für das Scrollen
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyDown: // Nach rechts scrollen
			scrollY++
		case tcell.KeyUp: // Nach links scrollen
			if scrollY > 0 {
				scrollY--
			}
		case tcell.KeyEscape: // Beenden
			return nil
		default:
			break
		}

		textView.ScrollTo(0, scrollY)

		return event
	})

	return textView
}

func scrollableTextView() *tview.TextView {
	// Daten für die horizontale Liste
	items := []string{
		"pod/frontend-6cc49dfb67-vnz4k", "pod/frontend-ajfhuiu32s-hfj3", "pod/frontend-ajfhuiu32s-hfj3", "pod/frontend-ajfhuiu32s-hfj3", "pod/frontend-ajfhuiu32s-hfj3",
	}

	// Kombiniere die Items in einer Zeile
	listContent := ""
	for _, item := range items {
		listContent += item + "   " // Trennzeichen für Sichtbarkeit
	}

	// Erstelle ein neues TextView
	textView := tview.NewTextView().
		SetText(listContent).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	textView.SetBorder(true).SetTitle("Pods").SetTitleAlign(tview.AlignLeft)

	// Initiale Scroll-Position
	scrollX := 0

	// Ereignis-Handling für das Scrollen
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyRight: // Nach rechts scrollen
			scrollX++
		case tcell.KeyLeft: // Nach links scrollen
			if scrollX > 0 {
				scrollX--
			}
		case tcell.KeyEscape: // Beenden
			return nil
		default:
			break
		}

		textView.ScrollTo(scrollX, 0)

		return event
	})

	return textView
}
