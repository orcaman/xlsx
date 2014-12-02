package xlsx

import (
	"bytes"
	"encoding/xml"

	. "gopkg.in/check.v1"
)

type SheetSuite struct{}

var _ = Suite(&SheetSuite{})

// Test we can add a Row to a Sheet
func (s *SheetSuite) TestAddRow(c *C) {
	var f *File
	f = NewFile()
	sheet := f.AddSheet("MySheet")
	row := sheet.AddRow()
	c.Assert(row, NotNil)
	c.Assert(len(sheet.Rows), Equals, 1)
}

func (s *SheetSuite) TestMakeXLSXSheetFromRows(c *C) {
	file := NewFile()
	sheet := file.AddSheet("Sheet1")
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.Value = "A cell!"
	refTable := NewSharedStringRefTable()
	styles := &xlsxStyleSheet{}
	xSheet := sheet.makeXLSXSheet(refTable, styles)
	c.Assert(xSheet.Dimension.Ref, Equals, "A1:A1")
	c.Assert(xSheet.SheetData.Row, HasLen, 1)
	xRow := xSheet.SheetData.Row[0]
	c.Assert(xRow.R, Equals, 1)
	c.Assert(xRow.Spans, Equals, "")
	c.Assert(xRow.C, HasLen, 1)
	xC := xRow.C[0]
	c.Assert(xC.R, Equals, "A1")
	c.Assert(xC.S, Equals, 0)
	c.Assert(xC.T, Equals, "s") // Shared string type
	c.Assert(xC.V, Equals, "0") // reference to shared string
	xSST := refTable.makeXLSXSST()
	c.Assert(xSST.Count, Equals, 1)
	c.Assert(xSST.UniqueCount, Equals, 1)
	c.Assert(xSST.SI, HasLen, 1)
	xSI := xSST.SI[0]
	c.Assert(xSI.T, Equals, "A cell!")
}

// When we create the xlsxSheet we also populate the xlsxStyles struct
// with style information.
func (s *SheetSuite) TestMakeXLSXSheetAlsoPopulatesXLSXSTyles(c *C) {
	file := NewFile()
	sheet := file.AddSheet("Sheet1")
	row := sheet.AddRow()

	cell1 := row.AddCell()
	cell1.Value = "A cell!"
	style1 := *NewStyle()
	style1.Font = *NewFont(10, "Verdana")
	style1.Fill = *NewFill("solid", "FFFFFFFF", "00000000")
	style1.Border = *NewBorder("none", "thin", "none", "thin")
	cell1.SetStyle(style1)

	// We need a second style to check that Xfs are populated correctly.
	cell2 := row.AddCell()
	cell2.Value = "Another cell!"
	style2 := *NewStyle()
	style2.Font = *NewFont(10, "Verdana")
	style2.Fill = *NewFill("solid", "FFFFFFFF", "00000000")
	style2.Border = *NewBorder("none", "thin", "none", "thin")
	cell2.SetStyle(style2)

	refTable := NewSharedStringRefTable()
	styles := &xlsxStyleSheet{}
	worksheet := sheet.makeXLSXSheet(refTable, styles)

	c.Assert(styles.Fonts.Count, Equals, 2)
	c.Assert(styles.Fonts.Font[0].Sz.Val, Equals, "10")
	c.Assert(styles.Fonts.Font[0].Name.Val, Equals, "Verdana")

	c.Assert(styles.Fills.Count, Equals, 2)
	c.Assert(styles.Fills.Fill[0].PatternFill.PatternType, Equals, "solid")
	c.Assert(styles.Fills.Fill[0].PatternFill.FgColor.RGB, Equals, "FFFFFFFF")
	c.Assert(styles.Fills.Fill[0].PatternFill.BgColor.RGB, Equals, "00000000")

	c.Assert(styles.Borders.Count, Equals, 2)
	c.Assert(styles.Borders.Border[0].Left.Style, Equals, "none")
	c.Assert(styles.Borders.Border[0].Right.Style, Equals, "thin")
	c.Assert(styles.Borders.Border[0].Top.Style, Equals, "none")
	c.Assert(styles.Borders.Border[0].Bottom.Style, Equals, "thin")

	c.Assert(styles.CellStyleXfs.Count, Equals, 2)
	// The 0th CellStyleXf could just be getting the zero values by default
	c.Assert(styles.CellStyleXfs.Xf[0].FontId, Equals, 0)
	c.Assert(styles.CellStyleXfs.Xf[0].FillId, Equals, 0)
	c.Assert(styles.CellStyleXfs.Xf[0].BorderId, Equals, 0)
	// The 1st element cannot get initialised this way by accident.
	c.Assert(styles.CellStyleXfs.Xf[1].FontId, Equals, 1)
	c.Assert(styles.CellStyleXfs.Xf[1].FillId, Equals, 1)
	c.Assert(styles.CellStyleXfs.Xf[1].BorderId, Equals, 1)

	c.Assert(styles.CellXfs.Count, Equals, 2)
	c.Assert(styles.CellXfs.Xf[0].FontId, Equals, 0)
	c.Assert(styles.CellXfs.Xf[0].FillId, Equals, 0)
	c.Assert(styles.CellXfs.Xf[0].BorderId, Equals, 0)
	// As above, we need the 1st element to make the test fail
	// when it should.
	c.Assert(styles.CellXfs.Xf[1].FontId, Equals, 1)
	c.Assert(styles.CellXfs.Xf[1].FillId, Equals, 1)
	c.Assert(styles.CellXfs.Xf[1].BorderId, Equals, 1)

	// Finally we check that the cell points to the right CellXf /
	// CellStyleXf.
	c.Assert(worksheet.SheetData.Row[0].C[0].S, Equals, 0)
	c.Assert(worksheet.SheetData.Row[0].C[1].S, Equals, 1)
}

func (s *SheetSuite) TestMarshalSheet(c *C) {
	file := NewFile()
	sheet := file.AddSheet("Sheet1")
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.Value = "A cell!"
	refTable := NewSharedStringRefTable()
	styles := &xlsxStyleSheet{}
	xSheet := sheet.makeXLSXSheet(refTable, styles)

	output := bytes.NewBufferString(xml.Header)
	body, err := xml.MarshalIndent(xSheet, "  ", "  ")
	c.Assert(err, IsNil)
	c.Assert(body, NotNil)
	_, err = output.Write(body)
	c.Assert(err, IsNil)
	expectedXLSXSheet := `<?xml version="1.0" encoding="UTF-8"?>
  <worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
    <dimension ref="A1:A1"></dimension>
    <cols>
      <col min="1" max="1" width="9.5"></col>
    </cols>
    <sheetData>
      <row r="1">
        <c r="A1" s="0" t="s">
          <v>0</v>
        </c>
      </row>
    </sheetData>
  </worksheet>`
	c.Assert(output.String(), Equals, expectedXLSXSheet)
}

func (s *SheetSuite) TestMarshalSheetWithMultipleCells(c *C) {
	file := NewFile()
	sheet := file.AddSheet("Sheet1")
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.Value = "A cell (with value 1)!"
	cell = row.AddCell()
	cell.Value = "A cell (with value 2)!"
	refTable := NewSharedStringRefTable()
	styles := &xlsxStyleSheet{}
	xSheet := sheet.makeXLSXSheet(refTable, styles)

	output := bytes.NewBufferString(xml.Header)
	body, err := xml.MarshalIndent(xSheet, "  ", "  ")
	c.Assert(err, IsNil)
	c.Assert(body, NotNil)
	_, err = output.Write(body)
	c.Assert(err, IsNil)
	expectedXLSXSheet := `<?xml version="1.0" encoding="UTF-8"?>
  <worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
    <dimension ref="A1:B1"></dimension>
    <cols>
      <col min="1" max="1" width="9.5"></col>
      <col min="2" max="2" width="9.5"></col>
    </cols>
    <sheetData>
      <row r="1">
        <c r="A1" s="0" t="s">
          <v>0</v>
        </c>
        <c r="B1" s="1" t="s">
          <v>1</v>
        </c>
      </row>
    </sheetData>
  </worksheet>`
	c.Assert(output.String(), Equals, expectedXLSXSheet)
}
