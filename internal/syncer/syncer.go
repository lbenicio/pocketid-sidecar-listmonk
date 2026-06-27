package syncer

import (
	"fmt"
	"log"

	"github.com/lbenicio/pocketid-sidecar-listmonk/internal/config"
	"github.com/lbenicio/pocketid-sidecar-listmonk/internal/listmonk"
	"github.com/lbenicio/pocketid-sidecar-listmonk/internal/pocketid"
)

// Stats holds counters for the sync operation.
type Stats struct {
	Created int
	Updated int
	Deleted int
	Errors  int
}

// Syncer reconciles PocketID users with a Listmonk list.
type Syncer struct {
	cfg       *config.Config
	pocketID  *pocketid.Client
	listmonk  *listmonk.Client
}

// New returns a new Syncer.
func New(cfg *config.Config) *Syncer {
	return &Syncer{
		cfg:      cfg,
		pocketID: pocketid.NewClient(cfg.PocketIDBaseURL, cfg.PocketIDAPIKey),
		listmonk: listmonk.NewClient(cfg.ListmonkBaseURL, cfg.ListmonkUsername, cfg.ListmonkPassword),
	}
}

// Run executes a full sync cycle.
func (s *Syncer) Run() (Stats, error) {
	stats := Stats{}

	log.Println("[sync] fetching PocketID users...")
	pocketUsers, err := s.pocketID.ListUsers()
	if err != nil {
		return stats, fmt.Errorf("fetching pocketid users: %w", err)
	}
	log.Printf("[sync] found %d users in PocketID", len(pocketUsers))

	log.Println("[sync] fetching Listmonk subscribers for list...")
	lmSubs, err := s.listmonk.ListSubscribersByList(s.cfg.ListmonkListID)
	if err != nil {
		return stats, fmt.Errorf("fetching listmonk subscribers: %w", err)
	}
	log.Printf("[sync] found %d subscribers in Listmonk list %d", len(lmSubs), s.cfg.ListmonkListID)

	// Build a lookup: pocketid_id → listmonk subscriber
	lmByPocketID := make(map[string]listmonk.Subscriber)
	for _, sub := range lmSubs {
		if pid, ok := sub.Attributes["pocketid_id"].(string); ok && pid != "" {
			lmByPocketID[pid] = sub
		}
	}

	// Build a set of PocketID user IDs for deletion detection.
	pocketIDSet := make(map[string]struct{}, len(pocketUsers))
	for _, u := range pocketUsers {
		pocketIDSet[u.ID] = struct{}{}
	}

	// --- Create or update from PocketID → Listmonk ---
	for _, pu := range pocketUsers {
		displayName := buildDisplayName(pu.FirstName, pu.LastName, pu.Username)
		existing, exists := lmByPocketID[pu.ID]

		if !exists {
			// New user — create subscriber
			if s.cfg.DryRun {
				log.Printf("[dry-run] would CREATE  subscriber for pocketid=%s email=%s name=%s", pu.ID, pu.Email, displayName)
			} else {
				log.Printf("[sync] CREATE subscriber for pocketid=%s email=%s name=%s", pu.ID, pu.Email, displayName)
				if _, err := s.listmonk.CreateSubscriber(pu.Email, displayName, s.cfg.ListmonkListID, pu.ID); err != nil {
					log.Printf("[sync] ERROR creating subscriber: %v", err)
					stats.Errors++
				} else {
					stats.Created++
				}
			}
			continue
		}

		// Existing user — check if update is needed
		if existing.Email != pu.Email || existing.Name != displayName {
			if s.cfg.DryRun {
				log.Printf("[dry-run] would UPDATE  subscriber id=%d (pocketid=%s) email=%q→%q name=%q→%q",
					existing.ID, pu.ID, existing.Email, pu.Email, existing.Name, displayName)
			} else {
				log.Printf("[sync] UPDATE subscriber id=%d (pocketid=%s) email=%q→%q name=%q→%q",
					existing.ID, pu.ID, existing.Email, pu.Email, existing.Name, displayName)
				if _, err := s.listmonk.UpdateSubscriber(existing.ID, pu.Email, displayName); err != nil {
					log.Printf("[sync] ERROR updating subscriber %d: %v", existing.ID, err)
					stats.Errors++
				} else {
					stats.Updated++
				}
			}
		}
	}

	// --- Delete Listmonk subscribers not in PocketID ---
	for _, sub := range lmSubs {
		pid, ok := sub.Attributes["pocketid_id"].(string)
		if !ok || pid == "" {
			continue // skip subscribers not managed by this syncer
		}
		if _, exists := pocketIDSet[pid]; exists {
			continue // still in PocketID, keep it
		}

		if s.cfg.DryRun {
			log.Printf("[dry-run] would DELETE subscriber id=%d (pocketid=%s) email=%s name=%s", sub.ID, pid, sub.Email, sub.Name)
		} else {
			log.Printf("[sync] DELETE subscriber id=%d (pocketid=%s) email=%s name=%s", sub.ID, pid, sub.Email, sub.Name)
			if err := s.listmonk.DeleteSubscriber(sub.ID); err != nil {
				log.Printf("[sync] ERROR deleting subscriber %d: %v", sub.ID, err)
				stats.Errors++
			} else {
				stats.Deleted++
			}
		}
	}

	return stats, nil
}

// buildDisplayName constructs a human-readable name from the available fields.
func buildDisplayName(firstName, lastName, username string) string {
	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	}
	if firstName != "" {
		return firstName
	}
	if lastName != "" {
		return lastName
	}
	return username
}
