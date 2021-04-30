package skat

type BlindedPlayerState struct {
	Ncards        int  `json:"ncards"`
	SeedProvided  bool `json:"seedProvided"`
	WonCardPoints int  `json:"wonPoints"`
	AwardedScore  int  `json:"awardedScore"`
}

type BlindedBiddingState struct {
	LastBid          int  `json:"lastCall"`
	Caller           int  `json:"caller"`
	Responder        int  `json:"responder"`
	AwaitingResponse bool `json:"awaitingResponse"`
}

type BlindedGameState struct {
	Phase GamePhase `json:"phase"`

	// Common state
	Players    []BlindedPlayerState `json:"players"`
	Hand       CardSet              `json:"hand"`
	SkatCards  int                  `json:"skatCards"`
	ServerSeed Seed                 `json:"serverSeed"`

	// Bidding state
	BiddingState    *BlindedBiddingState
	Declarer        int `json:"declarer"`
	LastBiddingCall int `json:"lastBiddingCall"`

	// Playing state
	CurrentForehand    int          `json:"currentForehand"`
	CurrentPlayer      int          `json:"currentPlayer"`
	GameType           GameType     `json:"gameType"`
	AnnouncedModifiers GameModifier `json:"announcedModifiers"`
	Table              CardSet      `json:"table"`

	// Scored state
	LossReason     string       `json:"lossReason"`
	FinalModifiers GameModifier `json:"finalModifiers"`
	FinalGameValue int          `json:"finalGameValue"`
	JackStrength   int          `json:"jackStrength"`
}
