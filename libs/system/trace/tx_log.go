package trace

type txLog struct {
	startTime int64
	AllCost   int64
	Record    map[string]*operateInfo // not support pall tx
}

func newTxLog() *txLog {
	tmp := &txLog{
		startTime: getNowTimeMs(),
		Record:    make(map[string]*operateInfo),
	}

	return tmp
}

func (s *txLog) StartTxLog(oper string) {
	if _, ok := s.Record[oper]; !ok {
		s.Record[oper] = newOperateInfo()
	}
	s.Record[oper].StartOper()
}

func (s *txLog) StopTxLog(oper string) {
	if _, ok := s.Record[oper]; !ok {
		return
	}
	s.Record[oper].StopOper()
}
