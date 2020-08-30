package main

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func (n *SSVNode) Start() {
	//go func() {
	//	select {
	//	case slot := <- n.ticker.C():
	//
	//	}
	//}()
}

func (n *SSVNode) GetTaskStream(req *ethpb.StreamRequest, s ethpb.SSV_GetTaskStreamServer) error  {
	log.Printf("New task stream established")
	go func() {
		for {
			select {
			case <- n.ctx.Done():
				return
			case slot := <- n.ticker.C():
				task := &ethpb.SSVTask{
					PublicKey:            req.GetPublicKeys()[0], // TODO should not be hard coded
					Topic:                ethpb.StreamTopics_SIGN_ATTESTATION,
					Data:                 &ethpb.SSVTask_Attestation{
						Attestation:&ethpb.AttestationData{
							Slot:                 slot,
							CommitteeIndex:       1,
							BeaconBlockRoot:      make([]byte,32),
							Source:               &ethpb.Checkpoint{
								Epoch:                2,
								Root:                 make([]byte,32),
							},
							Target:               &ethpb.Checkpoint{
								Epoch:                3,
								Root:                 make([]byte,32),
							},
						},
					},
				}
				err := s.Send(task)
				if err != nil {
					log.WithError(err).Printf("could not stream task")
				} else {
					log.Printf("streamed task for slot: %d", slot)
				}
			}
		}
	}()

	<- n.ctx.Done()

	return nil
}

