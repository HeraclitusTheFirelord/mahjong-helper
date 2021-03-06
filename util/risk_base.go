package util

// 计算安牌以及可以视作筋牌的 123789 牌
func calcSujiSafeTiles27(safeTiles34 []bool, leftTiles34 []int) []int {
	sujiSafeTiles27 := make([]int, 27)
	const _true = 1
	for i, safe := range safeTiles34[:27] {
		if safe {
			sujiSafeTiles27[i] = _true
		}
	}
	for i := 0; i < 3; i++ {
		// 2断，当做打过1
		if leftTiles34[9*i+1] == 0 {
			sujiSafeTiles27[9*i] = _true
		}
		// 3断，当做打过12
		if leftTiles34[9*i+2] == 0 {
			sujiSafeTiles27[9*i] = _true
			sujiSafeTiles27[9*i+1] = _true
		}
		// 4断，当做打过23
		if leftTiles34[9*i+3] == 0 {
			sujiSafeTiles27[9*i+1] = _true
			sujiSafeTiles27[9*i+2] = _true
		}
		// 6断，当做打过78
		if leftTiles34[9*i+5] == 0 {
			sujiSafeTiles27[9*i+6] = _true
			sujiSafeTiles27[9*i+7] = _true
		}
		// 7断，当做打过89
		if leftTiles34[9*i+6] == 0 {
			sujiSafeTiles27[9*i+7] = _true
			sujiSafeTiles27[9*i+8] = _true
		}
		// 8断，当做打过9
		if leftTiles34[9*i+7] == 0 {
			sujiSafeTiles27[9*i+8] = _true
		}
	}
	return sujiSafeTiles27
}

type RiskTiles34 []float64

// 根据巡目（对于对手而言）、现物、立直后通过的牌、NC、Dora，来计算基础铳率
// 至于早外、OC 和读牌交给后续的计算
// turns: 巡目，这里是对于对手而言的，也就是该玩家舍牌的次数
// safeTiles34: 现物及立直后通过的牌
// leftTiles34: 各个牌在山中剩余的枚数
// roundWindTile: 场风
// playerWindTile: 自风
func CalculateRiskTiles34(turns int, safeTiles34 []bool, leftTiles34 []int, doraTiles []int, roundWindTile int, playerWindTile int) (risk34 RiskTiles34) {
	risk34 = make(RiskTiles34, 34)

	// 只对 dora 牌的危险度进行调整（综合了放铳率和失点）
	// double dora 等的危险度会进一步升高
	doraMulti := func(tile int, tileType tileType) float64 {
		multi := 1.0
		for _, dora := range doraTiles {
			if tile == dora {
				multi *= FixedDoraRiskRateMulti[tileType]
			}
		}
		return multi
	}

	// 生成用来计算筋牌的「安牌」
	sujiSafeTiles27 := calcSujiSafeTiles27(safeTiles34, leftTiles34)
	// 利用「安牌」计算无筋、筋、半筋、双筋的铳率
	// TODO: 单独处理宣言牌的筋牌、宣言牌的同色牌的铳率
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][sujiSafeTiles27[idx+3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			mixSafeTile := sujiSafeTiles27[idx-3]<<1 | sujiSafeTiles27[idx+3]
			t := TileTypeTable[j][mixSafeTile]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][sujiSafeTiles27[idx-3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
		// 5断，37视作安牌筋
		if leftTiles34[9*i+4] == 0 {
			t := tileTypeSuji37
			risk34[9*i+2] = RiskRate[turns][t] * doraMulti(9*i+2, t)
			risk34[9*i+6] = RiskRate[turns][t] * doraMulti(9*i+6, t)
		}
	}
	for i := 27; i < 34; i++ {
		if leftTiles34[i] > 0 {
			// 该玩家的役牌 = 场风/其自风/白/发/中
			isYakuHai := i == roundWindTile || i == playerWindTile || i >= 31
			t := HonorTileType[boolToInt(isYakuHai)][leftTiles34[i]-1]
			risk34[i] = RiskRate[turns][t] * doraMulti(i, t)
		} else {
			// 剩余数为0可以视作安牌（只输国士）
			risk34[i] = 0
		}
	}

	// 更新铳率表：NC牌的安牌
	// 12和筋1差不多（2比1多10%）
	// 3和筋2差不多
	// 456和两筋差不多（存疑？）
	ncSafeTile34 := CalcNCSafeTiles(leftTiles34)
	for _, ncSafeTile := range ncSafeTile34 {
		idx := ncSafeTile.Tile34
		switch idx % 9 {
		case 1, 9:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 2, 8:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * 1.1 * doraMulti(idx, t)
		case 3, 7:
			t := tileTypeSuji28
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 4, 6:
			t := tileTypeDoubleSuji46
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 5:
			t := tileTypeDoubleSuji5
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
	}

	// 更新铳率表：DNC且剩余枚数为0的也当作安牌（忽略国士）
	dncSafeTiles := CalcDNCSafeTiles(leftTiles34)
	for _, dncSafeTile := range dncSafeTiles {
		if leftTiles34[dncSafeTile.Tile34] == 0 {
			risk34[dncSafeTile.Tile34] = 0
		}
	}

	// 更新铳率表：现物的铳率为0
	for i, isSafe := range safeTiles34 {
		if isSafe {
			risk34[i] = 0
		}
	}

	return
}

// 对 5 巡前的外侧牌的危险度进行调整
// 粗略调整为 *0.5
func (l RiskTiles34) FixWithEarlyOutside(discardTiles []int) RiskTiles34 {
	for _, dTile := range discardTiles {
		l[dTile] *= 0.5
	}
	return l
}

func (l RiskTiles34) FixWithGlobalMulti(multi float64) RiskTiles34 {
	for i := range l {
		l[i] *= multi
	}
	return l
}

// 根据副露情况对危险度进行修正
func (l RiskTiles34) FixWithPoint(ronPoint float64) RiskTiles34 {
	return l.FixWithGlobalMulti(ronPoint / RonPointRiichiHiIppatsu)
}

// 计算剩余的无筋 123789 牌
// 总计 18 种。剩余无筋牌数量越少，该无筋牌越危险
func CalculateLeftNoSujiTiles(safeTiles34 []bool, leftTiles34 []int) (leftNoSujiTiles []int) {
	isNoSujiTiles27 := make([]bool, 27)

	for i := 0; i < 3; i++ {
		// 根据 456 中张是否为安牌来判断相应筋牌是否安全
		for j := 3; j < 6; j++ {
			if !safeTiles34[9*i+j] {
				isNoSujiTiles27[9*i+j-3] = true
				isNoSujiTiles27[9*i+j+3] = true
			}
		}
		// 5断，37视作安牌筋
		if leftTiles34[9*i+4] == 0 {
			isNoSujiTiles27[9*i+2] = false
			isNoSujiTiles27[9*i+6] = false
		}
	}

	// 根据打过 4 张的壁牌更新 isNoSujiTiles27
	for i, c := range leftTiles34[:27] {
		if c == 0 {
			isNoSujiTiles27[i] = false
		}
	}

	// 根据 No Chance 的安牌更新 isNoSujiTiles27
	sujiSafeTiles27 := calcSujiSafeTiles27(safeTiles34, leftTiles34)
	const _true = 1
	for i, isSafe := range sujiSafeTiles27 {
		if isSafe == _true {
			isNoSujiTiles27[i] = false
		}
	}

	for i, isNoSujiTile := range isNoSujiTiles27 {
		if isNoSujiTile {
			leftNoSujiTiles = append(leftNoSujiTiles, i)
		}
	}

	return
}

// TODO:（待定）有早外的半筋（早巡打过8m时，3m的半筋6m）
// TODO:（待定）利用赤宝牌计算铳率
// TODO: 宝牌周边牌的危险度要增加一点
