package babel

type PendingAcknowledgementRequest struct {
	acked func() error
	missed func() error

	timer 
	opaque uint16
}

func NewPendingAcknowledgementRequest(intv time.Duration, acked, missed func() error) PendingAcknowledgementRequest {
	pAck := PendingAcknowledgementRequest{

	}

	pAck.timer = NewTimeout(intv)

	return pack
}

type Speaker struct {
	CalculateMetric func(r Route, n Neighbour) uint16
	FilterRoute func(r Route) bool
	SelectRoute(rs []Route) []Route

	pendingAcknowledgementRequests AcknowledgementRequestTable
}

func (s *Speaker) run() {
	s.startUpdateTimer()
	s.startHelloTimer()

	for {
		pkt := s.read()

		if err := s.onPacket(pkt); err != nil {
			s.log.Error(“Failed to handle packet: %w”, err)
		}
	}
}

func (s *Speaker) onPacket(pkt) (err error) {
	var n *Neighbour
	var ackReq tlv.AcknowledgementRequest

	if n, ok := s.neighbors.Get(pkt.remote); !ok {
		n = NewNeighbour(pkt.remote)
		s.neighbors.Put(n)
	}

	for _, v in pkt.vs {
		if v, ok := v.(tlv.AcknowledgementRequest); ok {
			ackReq = v
		} else if err := s.onValue(v, n); err != nil {
			return err
		}
	}

	// Send acknowledgements only after all other TLVs have been successfully handled.
	if ackReq != nil {
		if err := s.handleAcknowledgementRequest(ackReq); err != nil {
			return err
		}
	}

	return nil
}

func (s *Speaker) onValue(v tlv.Value) error {
	switch v := v.(type) {
		case tlv.Pad1, tlv.PadN:
			return nil

		case tlv.Acknowledgement:
			return s.onAcknowledgement(v)

		case tlv.Hello:
			return s.onHello(v, pkt.remote.IsMulticast())

		case tlv.IHU:
			return s.onIHU(v)

		case tlv.Update:
			return s.onUpdate(v)

		case tlv.RouteRequest:
			return s.onRouteRequest(v)

		case tlv.SeqnoRequest:
			return s.onSeqnoRequest(v)

		default:
			s.log.Trace(“Received unsupported TLV: %#x” v.Type)
			return nil
		}
}

// Handlers

func (s *Speaker) onHello(h tlv.Hello, isMulticast bool) error {
	n.updateCosts()
	s.selectRoutes()
	n.resetHelloTimeout()

	return nil
}

func (s *Speaker) onIHU(ihr tlv.IHU) error {
	// TODO

	return nil
}

// Section 3.5.3 Route Acquisition
func (s *Speaker) onUpdate(upd tlv.Update, n *Neighbor) error {
	var current, next *Route

	if current, ok := s.routes.Get(upd.Prefix, n); !ok {
		if !s.isUpdateFeasible(upd) {
			s.log.Trace(“Ignoring infeasible update %s from neighbor %s”, upd, n)
			return
		}

		if upd.Metric == Inf {
			s.log.Trace(“Ignoring retraction update %s from neighbor %s”, upd, n)
			return
		}

		next = Route{
			RouterID: upd.RouterID,
			NextHop: upd.NextHop,
			Seqno: upd.Seqno,
			Metric: upd.Metric,
		}

		s.routes.Put(upd.Prefix, n, next)
	} else {
		if current.Selected && !feasible && upd.RouterID == current.RouterID {
			// Update MAY be ignored
		} else {
			next := *current

			next.Seqno = upd.Seqno
			next.Metric = upd.Metric
			next.RouterID = upd.RouterID

			if rte.Metric != Inf {
				rte.resetExpiryTimer()
			}

			if !s.isUpdateFeasible(upd) {

				// Section 3.8.2.2. Dealing with Unfeasible Updates
				if current.Selected {
					if err := n.sendSeqnoRequest(current.Prefix); err != nil {
						return fmt.Errorf(“failed to send seqno request for selected route: %w”, err)
					}
				}

				next.Selected = false
			}

			if next.RouterID != current.RouterID {
				// Send urgent update / retraction
			}

			s.routes.Update(&next)
		}
	}
}

func (s *Speaker) onRouteRequest(req tlv.RouteRequest) error {
	// TODO
} 

func (s *Speaker) onSeqnoRequest(req tlv.SeqnoRequest) error {
	// TODO
}

func (s *Speaker) onAcknowledgementRequest(req tlv.AcknowledgementRequst) error {
	if err := s.send([]tlv.Values{
		tlv.Acknowledgement{
			Opaque: req.Opaque
		},
	}); err != nil {
		return fmt.Errorf(“failed to handle acknowledgement request: %w”, err)
	}

	return nil
}

func (s *Speaker) onAcknowledgement(ack tlv.Acknowledgement) error {
	if pAck, ok := s.pendingAcknowledgementRequests.Pop(ack.Opaque); ok {
		if err := pAck.acked(); err != nil {
			return fmt.Errorf(“failed to acknowledge request: %w”, err)
		}

		if err := pAck.timeout.Close(); err != nil {
			return fmt.Errorf(“failed to close pending acknowledgment request timer: %w”, err)
		}
	} else {
		s.log.Warn(“Received unexpected acknowledgement”)
	}

	return nil
}

// Timers and timeouts

func (s *Speaker) onHelloTimer() {
	// TODO: Send periodic hellos
}

func (n *Neighbour) onNeighbourTimeout() {
	// TODO: Drop neighbour if history becomes empty?
}

func (n *Neighbour) onHelloTimeout() {
	// TODO
}

// Speaker

func (s *Speaker) onUpdateTimeout() {
	s.sendUpdatesFull()
}

func (s *Speaker) onIHUTimeout() {
	// TODO
}


// Section 3.6 Route Selection
func (s *Speaker) selectRoutes() error {
	updated := []Route{}

	for _, rs in s.routes.ForEachPrefix {
		var next, cur *Route

		next = s.selectRoute(rs)

		for _, r := range rs {
			if r.Selected {
				cur = r
			}

			r.Selected = r == next
		}

		if current == nil || s.triggersUpdate(next, current) {
			updated = append(updated, next)
		}
	}
}

func (s *Speaker) sendWithAcknowledgementRequest(vs []tlv.Values, intv time.Duration, acked, missed func() error) error {
	magic := math.Rand()

	ack := tlv.AcknowledgementRequest{
		Opaque: magic,
		Interval: into
	}

	vsa := []tlv.Values{
		ack,
	}
	vsa = append(vsa, vs…)

	pAck := NewPendingAcknowledgement(acked, missed, to)
	s.pendingAcknowledgementRequests.Put(magic, pAck)

	return s.send(vsa)
}

func (s *Speaker) sendMulticastHello() error {
	errs := []error{}

	hello := tlv.Hello{
		// TODO
	}

	for _, intf := range s.interfaces {
		if err := s.sendMulticast(intf, hello); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		// TODO: Return errgroup
	}

	return nil
}

func (s *Speaker) sendUpdatesTriggered(us []tlv.Update, urgent bool) error {
	return s.sendUpdate(us, urgent)
}

func (s *Speaker) sendUpdatesFull() error {
	us := []tlv.Update{}

	for _, r := range s.routes.Selected {
		us = append(us, r.Update())
	}

	return s.sendUpdates(us, false)
}

func (s *Speaker) sendUpdates(us []tlv.Update, urgent bool, ack bool) error {
	// TODO: Split-horizon

	sortUpdatesByPrefix(us)

	if !urgent {
		if ack {
			return 
			return s.send(us)
	}
	

	if s.neighbors.Length() > partialUpdateNeighborsAcknowledgementThreshold {
		for i := 0; i < 4; i++ {
			if err := s.sendUrgent(us); err != nil {
				return err
			}
		}
	} else {
		errs := []error{}

		for _, n := range s.neighbors {
			errs = append(errs, n.sendUrgentWithAcknowledgementRequest(upd))
		}

		if len(errs) > 0 {
			// TODO
		}
	}
}

func (s *Speaker) sendSeqnoRequest(pfx Prefix) error {

	return s.send([]tlv.Value{
		tlv.SeqnoRequest{
			// TODO
		}
	})

}

func (s *Speaker) isUpdateFeasible(upd tlv.Update) bool {
	if upd.Metric == Inf {
		// Retractions are always feasible
		return True
	}

	if src, ok := s.sources.Get(upd.Prefix, upd.RouterID); !ok {
		return frue
	} else if ok {
		return src.Distance.IsBetterThan(upd.Distance)
	}
	
	return false
}

func (s *Speaker) selectRoute(rs []*Route) *Route {
	var sr *Route

	for _, r := range rs {
		if sr == nil || r.Metric < sr.Metric {
			sr = r
		}
	}

	return sr
}

// Neighbour

func (n *Neighbour) updateCost() error {
	// TODO
}

func (n *Neighbour) sendUnicastHello() error {
	// TODO
}

func (n *Neighbour) resetHelloTimeout() {
	// TODO
}
