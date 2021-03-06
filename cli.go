package main

import (
	"fmt"
	"strings"
	"github.com/fatih/color"
	"sort"
	"github.com/EndlessCheng/mahjong-helper/util"
)

func printAccountInfo(accountID int) {
	fmt.Printf("您的账号 ID 为 ")
	color.New(color.FgHiGreen).Printf("%d", accountID)
	fmt.Printf("，该数字为雀魂服务器账号数据库中的 ID，该值越小表示您的注册时间越早\n")
}

//

type handsRisk struct {
	tile int
	risk float64
}

// 34 种牌的危险度
type riskTable util.RiskTiles34

func (t riskTable) printWithHands(hands []int, fixedRiskMulti float64) {
	// 打印铳率=0的牌（现物，或NC且剩余数=0）
	safeCount := 0
	for i, c := range hands {
		if c > 0 && t[i] == 0 {
			fmt.Printf(" " + util.MahjongZH[i])
			safeCount++
		}
	}

	// 打印危险牌，按照铳率排序&高亮
	handsRisks := []handsRisk{}
	for i, c := range hands {
		if c > 0 && t[i] > 0 {
			handsRisks = append(handsRisks, handsRisk{i, t[i]})
		}
	}
	sort.Slice(handsRisks, func(i, j int) bool {
		return handsRisks[i].risk < handsRisks[j].risk
	})
	if len(handsRisks) > 0 {
		if safeCount > 0 {
			fmt.Print(" |")
		}
		for _, hr := range handsRisks {
			// 颜色考虑了听牌率
			color.New(getNumRiskColor(hr.risk * fixedRiskMulti)).Printf(" " + util.MahjongZH[hr.tile])
		}
	}
}

type riskInfo struct {
	// 该玩家的听牌率（立直时为 100.0）
	tenpaiRate float64

	// 各种牌的铳率表
	riskTable riskTable

	// 剩余无筋 123789
	// 总计 18 种。剩余无筋牌数量越少，该无筋牌越危险
	leftNoSujiTiles []int

	// 荣和点数
	// 仅调试用
	_ronPoint float64
}

type riskInfoList []riskInfo

func (l riskInfoList) printWithHands(hands []int, leftCounts []int) {
	const tenpaiRateLimit = 50.0
	dangerousPlayerCount := 0
	// 打印安牌，危险牌
	names := []string{"", "下家", "对家", "上家"}
	for i := len(l) - 1; i >= 1; i-- {
		// 听牌率超过 50% 就打印铳率
		tenpaiRate := l[i].tenpaiRate
		if len(l[i].riskTable) > 0 && (debugMode || tenpaiRate > tenpaiRateLimit) {
			dangerousPlayerCount++
			fmt.Print(names[i] + "安牌:")
			//if debugMode {
			//fmt.Printf("(%d*%2.2f%%听牌率)", int(l[i]._ronPoint), l[i].tenpaiRate)
			//}
			l[i].riskTable.printWithHands(hands, tenpaiRate/100)

			// 打印无筋数量和种类
			noSujiInfo := util.TilesToStr(l[i].leftNoSujiTiles)
			if len(l[i].leftNoSujiTiles) == 0 {
				noSujiInfo = "非好型听牌/振听"
			}
			if !debugMode {
				fmt.Printf(" [%d无筋]", len(l[i].leftNoSujiTiles))
			} else {
				fmt.Printf(" [%d无筋: %s]", len(l[i].leftNoSujiTiles), noSujiInfo)
			}

			fmt.Println()
		}
	}

	// 若不止一个玩家立直/副露，打印加权综合铳率（考虑了听牌率）
	mixedPlayers := 0
	for _, ri := range l[1:] {
		if ri.tenpaiRate > 0 {
			mixedPlayers++
		}
	}
	if dangerousPlayerCount > 0 && mixedPlayers > 1 {
		fmt.Print("综合安牌:")
		mixedRiskTable := make(riskTable, 34)
		for i := range mixedRiskTable {
			mixedRisk := 0.0
			for _, ri := range l[1:] {
				_risk := ri.riskTable[i] * ri.tenpaiRate / 100
				mixedRisk = mixedRisk + _risk - mixedRisk*_risk/100
			}
			mixedRiskTable[i] = mixedRisk
		}
		mixedRiskTable.printWithHands(hands, 1)
		fmt.Println()
	}

	// 打印因 NC OC 产生的安牌
	// TODO: 重构至其他函数
	if dangerousPlayerCount > 0 {
		ncSafeTileList := util.CalcNCSafeTiles(leftCounts).FilterWithHands(hands)
		ocSafeTileList := util.CalcOCSafeTiles(leftCounts).FilterWithHands(hands)
		if len(ncSafeTileList) > 0 {
			fmt.Printf("NC:")
			for _, safeTile := range ncSafeTileList {
				fmt.Printf(" " + util.MahjongZH[safeTile.Tile34])
			}
			fmt.Println()
		}
		if len(ocSafeTileList) > 0 {
			fmt.Printf("OC:")
			for _, safeTile := range ocSafeTileList {
				fmt.Printf(" " + util.MahjongZH[safeTile.Tile34])
			}
			fmt.Println()
		}

		// 下面这个是另一种显示方式：显示壁牌
		//printedNC := false
		//for i, c := range leftCounts[:27] {
		//	if c != 0 || i%9 == 0 || i%9 == 8 {
		//		continue
		//	}
		//	if !printedNC {
		//		printedNC = true
		//		fmt.Printf("NC:")
		//	}
		//	fmt.Printf(" " + util.MahjongZH[i])
		//}
		//if printedNC {
		//	fmt.Println()
		//}
		//printedOC := false
		//for i, c := range leftCounts[:27] {
		//	if c != 1 || i%9 == 0 || i%9 == 8 {
		//		continue
		//	}
		//	if !printedOC {
		//		printedOC = true
		//		fmt.Printf("OC:")
		//	}
		//	fmt.Printf(" " + util.MahjongZH[i])
		//}
		//if printedOC {
		//	fmt.Println()
		//}
		fmt.Println()
	}
}

//

/*

8     切 3索 听[2万, 7万]
9.20  [20 改良]  4.00 听牌数

4     听 [2万, 7万]
4.50  [ 4 改良]  55.36% 参考和率

8     45万吃，切 4万 听[2万, 7万]
9.20  [20 改良]  4.00 听牌数

*/
// 打印何切分析结果（双行）
func printWaitsWithImproves13_twoRows(result13 *util.WaitsWithImproves13, discardTile34 int, openTiles34 []int) {
	shanten := result13.Shanten
	waits := result13.Waits

	waitsCount, waitTiles := waits.ParseIndex()
	c := getWaitsCountColor(shanten, float64(waitsCount))
	color.New(c).Printf("%-6d", waitsCount)
	if discardTile34 != -1 {
		if len(openTiles34) > 0 {
			meldType := "吃"
			if openTiles34[0] == openTiles34[1] {
				meldType = "碰"
			}
			color.New(color.FgHiWhite).Printf("%s%s", string([]rune(util.MahjongZH[openTiles34[0]])[:1]), util.MahjongZH[openTiles34[1]])
			fmt.Printf("%s，", meldType)
		}
		fmt.Print("切 ")
		if shanten <= 1 {
			color.New(getSelfDiscardRiskColor(discardTile34)).Print(util.MahjongZH[discardTile34])
		} else {
			fmt.Print(util.MahjongZH[discardTile34])
		}
		fmt.Print(" ")
	}
	//fmt.Print("等")
	//if shanten <= 1 {
	//	fmt.Print("[")
	//	if len(waitTiles) > 0 {
	//		fmt.Print(util.MahjongZH[waitTiles[0]])
	//		for _, idx := range waitTiles[1:] {
	//			fmt.Print(", " + util.MahjongZH[idx])
	//		}
	//	}
	//	fmt.Println("]")
	//} else {
	fmt.Println(util.TilesToStrWithBracket(waitTiles))
	//}

	if len(result13.Improves) > 0 {
		fmt.Printf("%-6.2f[%2d 改良]", result13.AvgImproveWaitsCount, len(result13.Improves))
	} else {
		fmt.Print(strings.Repeat(" ", 15))
	}

	fmt.Print(" ")

	if shanten >= 1 {
		c := getWaitsCountColor(shanten-1, result13.AvgNextShantenWaitsCount)
		color.New(c).Printf("%5.2f", result13.AvgNextShantenWaitsCount)
		fmt.Printf(" %s", util.NumberToChineseShanten(shanten-1))
		if shanten >= 2 {
			fmt.Printf("进张")
		} else { // shanten == 1
			fmt.Printf("数")
			if showAgariAboveShanten1 {
				fmt.Printf("（%.2f%% 参考和率）", result13.AvgAgariRate)
			}
		}
		if showScore {
			mixedScore := result13.MixedWaitsScore
			//for i := 2; i <= shanten; i++ {
			//	mixedScore /= 4
			//}
			fmt.Printf("（%.2f 综合分）", mixedScore)
		}
	} else { // shanten == 0
		fmt.Printf("%5.2f%% 参考和率", result13.AvgAgariRate)
	}

	//if dangerous {
	//	// TODO: 提示危险度！
	//}

	fmt.Println()
}

/*
4[ 4.56] 切 8饼 => 44.50% 参考和率[ 4 改良] [7p 7s] [振听]

31[33.58] 切7索 => 5.23听牌数 [16改良] [6789p 56789s] [可能振听]

48[50.64] 切5饼 => 24.25一向听进张 [12改良] [123456789p 56789s]

31[33.62] 77索碰,切5饼 => 5.48听牌数 [15 改良] [123456789p]

*/
// 打印何切分析结果（单行）
func printWaitsWithImproves13_oneRow(result13 *util.WaitsWithImproves13, discardTile34 int, openTiles34 []int) {
	shanten := result13.Shanten

	// 打印进张数
	waitsCount, waitTiles := result13.Waits.ParseIndex()
	c := getWaitsCountColor(shanten, float64(waitsCount))
	color.New(c).Printf("%2d", waitsCount)
	// 打印改良进张均值
	if len(result13.Improves) > 0 {
		fmt.Printf("[%5.2f]", result13.AvgImproveWaitsCount)
	} else {
		fmt.Print(strings.Repeat(" ", 7))
	}

	fmt.Print(" ")

	// 是否为3k+2张牌的何切分析
	if discardTile34 != -1 {
		// 打印鸣牌分析
		if len(openTiles34) > 0 {
			meldType := "吃"
			if openTiles34[0] == openTiles34[1] {
				meldType = "碰"
			}
			color.New(color.FgHiWhite).Printf("%s%s", string([]rune(util.MahjongZH[openTiles34[0]])[:1]), util.MahjongZH[openTiles34[1]])
			fmt.Printf("%s,", meldType)
		}
		// 打印舍牌
		fmt.Print("切")
		tileZH := util.MahjongZH[discardTile34]
		if discardTile34 >= 27 {
			tileZH = " " + tileZH
		}
		if shanten <= 1 {
			color.New(getSelfDiscardRiskColor(discardTile34)).Print(tileZH)
		} else {
			fmt.Print(tileZH)
		}
	}

	fmt.Print(" => ")

	if shanten >= 1 {
		// 打印前进后的进张数均值
		c := getWaitsCountColor(shanten-1, result13.AvgNextShantenWaitsCount)
		color.New(c).Printf("%5.2f", result13.AvgNextShantenWaitsCount)
		fmt.Printf("%s", util.NumberToChineseShanten(shanten-1))
		if shanten >= 2 {
			//fmt.Printf("进张")
		} else { // shanten == 1
			fmt.Printf("数")
			if showAgariAboveShanten1 {
				fmt.Printf("（%.2f%% 参考和率）", result13.AvgAgariRate)
			}
		}
		if showScore {
			mixedScore := result13.MixedWaitsScore
			//for i := 2; i <= shanten; i++ {
			//	mixedScore /= 4
			//}
			fmt.Printf("（%.2f 综合分）", mixedScore)
		}
	} else { // shanten == 0
		fmt.Printf("%5.2f%% 参考和率", result13.AvgAgariRate)
	}

	//if dangerous {
	//	// TODO: 提示危险度！
	//}

	fmt.Print(" ")

	// 打印改良数
	if len(result13.Improves) > 0 {
		fmt.Printf("[%2d改良]", len(result13.Improves))
	} else {
		fmt.Print(strings.Repeat(" ", 4))
		fmt.Print(strings.Repeat("　", 2)) // 全角空格
	}

	fmt.Print(" ")

	// 打印进张类型
	fmt.Print(util.TilesToStrWithBracket(waitTiles))

	// 打印振听提示
	if result13.FuritenRate > 0 {
		fmt.Print(" ")
		if result13.FuritenRate < 1 {
			color.New(color.FgHiYellow).Printf("[可能振听]")
		} else {
			color.New(color.FgHiRed).Printf("[振听]")
		}
	}

	fmt.Println()
}
