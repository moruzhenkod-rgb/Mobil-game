package main

// AdMob integration stubs.
//
// On desktop (Windows): no ads are shown; reward is granted directly.
// On Android/iOS (via gomobile): replace these stubs with the real
// AdMob SDK calls (github.com/unity-go/admob or similar bridge).

// AdState tracks the rewarded-ad lifecycle.
type AdState int

const (
	AdIdle     AdState = iota
	AdLoading          // ad request sent
	AdReady            // ad loaded and ready to show
	AdShowing          // currently displayed
	AdFinished         // user watched to completion
	AdFailed           // load or show error
)

type AdManager struct {
	state       AdState
	rewardReady bool
	onRewarded  func() // called when user earns the reward
}

var ads = &AdManager{}

// LoadRewardedAd requests a rewarded ad from AdMob.
// On desktop this is a no-op; ad is immediately "ready".
func (a *AdManager) LoadRewardedAd() {
	a.state = AdLoading
	// Simulate instant load on desktop
	a.state = AdReady
	a.rewardReady = true
}

// ShowRewardedAd presents the ad and calls onRewarded after completion.
func (a *AdManager) ShowRewardedAd(onRewarded func()) {
	if !a.rewardReady {
		return
	}
	a.onRewarded = onRewarded
	a.state = AdShowing

	// Desktop stub: grant reward immediately
	a.grantReward()
}

func (a *AdManager) grantReward() {
	a.state = AdFinished
	a.rewardReady = false
	if a.onRewarded != nil {
		a.onRewarded()
		a.onRewarded = nil
	}
	// Pre-load next ad
	a.LoadRewardedAd()
}

func (a *AdManager) IsReady() bool { return a.rewardReady && !progress.AdsRemoved }

// ShowBannerAd — stub for banner ads.
// In production, rendered natively by the AdMob SDK overlay.
func ShowBannerAd() {
	if progress.AdsRemoved {
		return
	}
	// No-op on desktop
}

// HideBannerAd hides the banner (e.g. during gameplay).
func HideBannerAd() {}

// TrackAdImpression logs an ad impression to Firebase Analytics.
func TrackAdImpression(adType string) {
	fb.LogEvent("ad_impression", map[string]interface{}{"ad_format": adType})
}
