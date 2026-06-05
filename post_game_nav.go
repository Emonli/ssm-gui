package main

import (
	"context"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"time"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/controllers"
	"github.com/kvarenzn/ssm/gui"
	"github.com/kvarenzn/ssm/log"
)

// ─────────────────────────────────────────────
// Post-Game Navigation (游戏结束后的导航)
// ─────────────────────────────────────────────
//
// After the game ends and autoplay completes, this module
// handles navigating through result screen, confirmation dialogs,
// and returning to song selection for the next game..
func postGameNavigationBanG(
	ctx context.Context,
	ADBdevice *adb.Device,
	sc *controllers.ScrcpyController,
	srv *gui.Server,
	ocrC *goOCRClient,
) bool {
	const (
		stageRankUp     = "RANK_UP"      //角色rank up弹窗 & 每日演出报酬弹窗
		stageOKButton   = "OK_BUTTON"    // 检查获得报酬弹窗OK按钮
		stagePopUpCheck = "POP_UP_CHECK" // 检查是否是一级弹窗，检查关闭按钮
		stagePlayAgain  = "PLAY_AGAIN"   // 返回歌曲选择
		stageContinue   = "CONTINUE"     // 点击继续/确定按钮
		maxIter         = 120            // 最多循环120次
	)

	log.Infoln("you are in post game navigation now")

	sampleFrameLuma := func(f controllers.ScrcpyFrame) float64 {
		if f.Width <= 0 || f.Height <= 0 || len(f.Plane0) < f.Width*f.Height {
			return 0
		}
		step := max(1, f.Width*f.Height/1024)
		var sum int64
		count := 0
		for i := 0; i < len(f.Plane0); i += step {
			sum += int64(f.Plane0[i])
			count++
		}
		if count == 0 {
			return 0
		}
		return float64(sum) / float64(count)
	}

	lastUIKey := ""
	lastUILogAt := time.Time{}
	lastLogStage := ""
	emitStage := func(stageName, sceneName, action string, screenLuma float64, force bool) {
		msg := fmt.Sprintf("%s\n  %s", stageName, action)
		uiKey := fmt.Sprintf("%s|%s|%s", stageName, sceneName, action)
		if force || uiKey != lastUIKey || time.Since(lastUILogAt) >= 700*time.Millisecond {
			srv.SetAutoTriggerDebug(gui.AutoTriggerDebug{
				Enabled:    true,
				Mode:       "bang",
				NavStage:   stageName,
				NavScene:   sceneName,
				ScreenLuma: screenLuma,
				Message:    msg,
			})
			lastUIKey = uiKey
			lastUILogAt = time.Now()
		}
		if force || stageName != lastLogStage {
			log.Infof("[NAV] %s\n", action)
			lastLogStage = stageName
		}
	}

	tapAt := func(x, y int) {
		ADBdevice.RawSh("input", "tap", strconv.Itoa(x), strconv.Itoa(y)) //nolint
		log.Infof("[POST_GAME_NAV] tap (%d,%d)\n", x, y)
		emitStage("click", "action", fmt.Sprintf("→ click (%d, %d)", x, y), 0, true)
	}

	stageName := stageRankUp
	stageEnteredAt := time.Now()
	setStage := func(next string) {
		if stageName == next {
			return
		}
		stageName = next
		stageEnteredAt = time.Now()
	}

	lastTapAt := time.Time{}
	_ = lastTapAt // reserved: may use for tap-rate limiting later
	const preStartActionDelay = 2 * time.Second
	const inActionDelay = 500 * time.Millisecond

	RankUpROI := [4]float64{roiRankUpButton.x1, roiRankUpButton.y1, roiRankUpButton.x2, roiRankUpButton.y2}
	ContinueButtonROI := [4]float64{roiContinueButton.x1, roiContinueButton.y1, roiContinueButton.x2, roiContinueButton.y2}
	CloseButtonROI := [4]float64{roiCloseButton.x1, roiCloseButton.y1, roiCloseButton.x2, roiCloseButton.y2}
	OKButtonROI := [4]float64{roiOKButton.x1, roiOKButton.y1, roiOKButton.x2, roiOKButton.y2}
	PlayAgainROI := [4]float64{roiPlayAgainButton.x1, roiPlayAgainButton.y1, roiPlayAgainButton.x2, roiPlayAgainButton.y2}

	RankUpKeywords := []string{"确定", "OK"}
	closeKeywords := []string{"关闭", "閉じる", "閉じ"}
	okKeywords := []string{"OK"}
	playAgainKeywords := []string{"再次演出", "もう1回ライブ"}
	continueKeywords := []string{"下一步", "次へ"}

	// wait for loading result screen
	time.Sleep(20 * time.Second)

	// main loop
	for i := 0; i < maxIter; i++ {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		frame, ok := sc.LatestFrame()
		if !ok || frame.Width == 0 || len(frame.Plane0) < frame.Width*frame.Height {
			emitStage(stageName, "", "→ waiting decoded frame", 0, false)
			time.Sleep(220 * time.Millisecond)
			continue
		}

		screenLuma := sampleFrameLuma(frame)

		switch stageName {
		case stageRankUp:
			// 检查角色升级弹窗（确定按钮）
			emitStage(stageRankUp, "rank-up-window-check", "→ 检查角色升级弹窗 确定/OK 按钮", screenLuma, false)
			pngData, scErr := ADBdevice.ScreencapPNGBytes()
			if scErr != nil {
				srv.SetError("screencap failed: " + scErr.Error())
				return false
			}
			rankUpTexts, ocrErr := ocrC.OCR(pngData, &RankUpROI)
			if ocrErr != nil {
				emitStage(stageRankUp, "rank-up-window-check", "→ OCR失敗，跳過本次檢查", screenLuma, false)
				continue
			}

			norm := normalizeSceneTexts(rankUpTexts)
			score := bestKeywordScore(norm, RankUpKeywords)
			emitStage(stageRankUp, "rank-up-window", fmt.Sprintf("→ OCR文字=%v | 相似度=%.2f", norm, score), screenLuma, false)

			if score >= 0.5 {
				emitStage(stageRankUp, "rank-up-window", "→ 检测到『确定/OK』按钮,准备点击", screenLuma, true)
				x, y, found := roiCenterPx(frame, roiRankUpButton)
				if found {
					tapAt(x, y)
					lastTapAt = time.Now()
					setStage(stageRankUp)
					time.Sleep(preStartActionDelay)
				}
				continue
			}
			time.Sleep(inActionDelay)
			if time.Since(stageEnteredAt) >= 900*time.Millisecond {
				setStage(stageOKButton)
			}

		case stageOKButton:
			// 检查二级弹窗（OK按钮）
			emitStage(stageOKButton, "sec-pop-up-window-check", "→ 检查二级弹窗 OK 按钮", screenLuma, false)
			pngData, scErr := ADBdevice.ScreencapPNGBytes()
			if scErr != nil {
				srv.SetError("screencap failed: " + scErr.Error())
				return false
			}
			okTexts, ocrErr := ocrC.OCR(pngData, &OKButtonROI)
			if ocrErr != nil {
				emitStage(stageOKButton, "sec-pop-up-window-check", "→ OCR失敗，跳過本次檢查", screenLuma, false)
				continue
			}

			norm := normalizeSceneTexts(okTexts)
			score := bestKeywordScore(norm, okKeywords)
			emitStage(stageOKButton, "sec-pop-up-window", fmt.Sprintf("→ OCR文字=%v | 相似度=%.2f", norm, score), screenLuma, false)

			if score >= 0.7 {
				emitStage(stageOKButton, "sec-pop-up-window", "→ 检测到『OK』按钮,准备点击", screenLuma, true)
				x, y, found := roiCenterPx(frame, roiOKButton)
				if found {
					tapAt(x, y)
					lastTapAt = time.Now()
					setStage(stagePopUpCheck)
					time.Sleep(preStartActionDelay)
				}
				continue
			}
			time.Sleep(inActionDelay)
			if time.Since(stageEnteredAt) >= 900*time.Millisecond {
				setStage(stagePopUpCheck)
			}

		case stagePopUpCheck:
			// 检查一级弹窗（关闭按钮）
			emitStage(stagePopUpCheck, "pop-up-window-check", "→ 检查一级弹窗 关闭/閉じる 按钮", screenLuma, false)
			pngData, scErr := ADBdevice.ScreencapPNGBytes()
			if scErr != nil {
				srv.SetError("screencap failed: " + scErr.Error())
				return false
			}
			closeTexts, ocrErr := ocrC.OCR(pngData, &CloseButtonROI)
			if ocrErr != nil {
				emitStage(stagePopUpCheck, "pop-up-window-check", "→ OCR失敗，跳過本次檢查", screenLuma, false)
				continue
			}

			norm := normalizeSceneTexts(closeTexts)
			score := bestKeywordScore(norm, closeKeywords)
			emitStage(stagePopUpCheck, "pop-up-window", fmt.Sprintf("→ OCR文字=%v | 相似度=%.2f", norm, score), screenLuma, false)

			if score >= 0.4 {
				emitStage(stagePopUpCheck, "pop-up-window", "→ 检测到『关闭/閉じる』按钮,准备点击", screenLuma, true)
				x, y, found := roiCenterPx(frame, roiCloseButton)
				if found {
					tapAt(x, y)
					lastTapAt = time.Now()
					setStage(stagePlayAgain)
					time.Sleep(preStartActionDelay)
				}
				continue
			}
			time.Sleep(inActionDelay)
			if time.Since(stageEnteredAt) >= 900*time.Millisecond {
				setStage(stagePlayAgain)
			}

		case stagePlayAgain:
			// 检查再次演出按钮
			emitStage(stagePlayAgain, "play-again-button-check", "→ 检测再次演出按钮", screenLuma, false)
			pngData, scErr := ADBdevice.ScreencapPNGBytes()
			if scErr != nil {
				srv.SetError("screencap failed: " + scErr.Error())
				return false
			}
			playAgainTexts, ocrErr := ocrC.OCR(pngData, &PlayAgainROI)
			if ocrErr != nil {
				emitStage(stagePlayAgain, "play-again-button-check", "→ OCR失敗，跳過本次檢查", screenLuma, false)
				continue
			}

			norm := normalizeSceneTexts(playAgainTexts)
			score := bestKeywordScore(norm, playAgainKeywords)
			emitStage(stagePlayAgain, "play-again-button", fmt.Sprintf("→ OCR文字=%v | 相似度=%.2f", norm, score), screenLuma, false)

			if score >= 0.4 {
				// 额外检查：按钮区域亮度 > 50% 才点击（避免灰化按钮误触发）
				roiLuma, _ := sampleROILuma(frame, roiPlayAgainButton)
				brightnessPct := roiLuma / 255.0 * 100
				emitStage(stagePlayAgain, "play-again-button", fmt.Sprintf("→ 按钮区域亮度=%.0f%%", brightnessPct), screenLuma, false)
				if brightnessPct >= 50 {
					emitStage(stagePlayAgain, "play-again-button", "→ 检测到『再次演出/もう1回ライブ』按钮,准备点击", screenLuma, true)
					x, y, found := roiCenterPx(frame, roiPlayAgainButton)
					if found {
						tapAt(x, y)
						lastTapAt = time.Now()
						emitStage(stagePlayAgain, "play-again-button", "→ 已点击,等待5秒后返回主近程", screenLuma, true)
						time.Sleep(5 * time.Second)
						return true
					}
					continue
				}
				emitStage(stagePlayAgain, "play-again-button", "→ 亮度不足，跳过", screenLuma, false)
			}
			time.Sleep(inActionDelay)
			if time.Since(stageEnteredAt) >= 900*time.Millisecond {
				setStage(stageContinue)
			}

		case stageContinue:
			// 检查下一步按钮
			emitStage(stageContinue, "continue-button-check", "→ 检测 下一步/次へ 按钮", screenLuma, false)
			pngData, scErr := ADBdevice.ScreencapPNGBytes()
			if scErr != nil {
				srv.SetError("screencap failed: " + scErr.Error())
				return false
			}
			continueTexts, ocrErr := ocrC.OCR(pngData, &ContinueButtonROI)
			if ocrErr != nil {
				emitStage(stageContinue, "continue-button-check", "→ OCR失敗，跳過本次檢查", screenLuma, false)
				continue
			}

			norm := normalizeSceneTexts(continueTexts)
			score := bestKeywordScore(norm, continueKeywords)
			emitStage(stageContinue, "continue-button", fmt.Sprintf("→ OCR文字=%v | 相似度=%.2f", norm, score), screenLuma, false)

			if score >= 0.4 {
				// 额外检查：按钮区域亮度 > 40% 才点击（避免灰化按钮误触发）
				roiLuma, _ := sampleROILuma(frame, roiContinueButton)
				brightnessPct := roiLuma / 255.0 * 100
				emitStage(stageContinue, "continue-button", fmt.Sprintf("→ 按钮区域亮度=%.0f%%", brightnessPct), screenLuma, false)
				if brightnessPct >= 40 {
					emitStage(stageContinue, "continue-button", "→ 检测到『下一步/次へ』按钮,准备点击", screenLuma, true)
					x, y, found := roiCenterPx(frame, roiContinueButton)
					if found {
						tapAt(x, y)
						lastTapAt = time.Now()
						setStage(stageRankUp)
						time.Sleep(preStartActionDelay)
					}
					continue
				}
				emitStage(stageContinue, "continue-button", "→ 亮度不足，跳过", screenLuma, false)
			}
			time.Sleep(inActionDelay)
			if time.Since(stageEnteredAt) >= 900*time.Millisecond {
				setStage(stageRankUp)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	emitStage(stagePlayAgain, "timeout", "→ post-game navigation timed out", 0, true)
	return false
}
