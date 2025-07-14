package blazego

func GetBlazeURL(urlType string) string {
	switch urlType {
	case "games":
		return "wss://api-gaming.blaze.bet.br/replication/?EIO=3&transport=websocket"
	case "general":
		return "wss://api-v2.blaze.bet.br/replication/?EIO=3&transport=websocket"
	default:
		return ""
	}
}
