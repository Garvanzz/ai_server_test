package define

const (
	ServerStateNormal = iota //0：正常
	ServerStateYongji        //1：拥挤
	ServerStateBaoMan        //爆满
	ServerStateWeihu         //维护
	ServerStateNoOpen        //未开服
	ServerStateStop          //停服

)
