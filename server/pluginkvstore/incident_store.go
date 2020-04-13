package pluginkvstore

import (
	"errors"
	"fmt"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-incident-response/server/incident"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	allHeadersKey = "all_headers"
	incidentKey   = "incident_"
)

type idHeaderMap map[string]incident.Header

// Ensure incidentStore implments the playbook.Store interface.
var _ incident.Store = (*incidentStore)(nil)

// incidentStore holds the information needed to fulfill the methods in the store interface.
type incidentStore struct {
	pluginAPI *pluginapi.Client
}

// NewIncidentStore creates a new store for incident ServiceImpl.
func NewIncidentStore(pluginAPI *pluginapi.Client) incident.Store {
	newStore := &incidentStore{
		pluginAPI: pluginAPI,
	}
	return newStore
}

// GetAllHeaders Gets all the header information.
func (s *incidentStore) GetHeaders(options incident.HeaderFilterOptions) ([]incident.Header, error) {
	headersMap, err := s.getIDHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get all headers value: %w", err)
	}

	headers := toHeader(headersMap)
	var result []incident.Header

	for _, header := range headers {
		if headerMatchesFilter(header, options) {
			result = append(result, header)
		}
	}

	return result, nil
}

// CreateIncident Creates a new incident.
func (s *incidentStore) CreateIncident(incdnt *incident.Incident) (*incident.Incident, error) {
	if incdnt == nil {
		return nil, errors.New("incident is nil")
	}
	if incdnt.ID != "" {
		return nil, errors.New("ID should not be set")
	}
	incdnt.ID = model.NewId()

	saved, err := s.pluginAPI.KV.Set(toIncidentKey(incdnt.ID), incdnt)
	if err != nil {
		return nil, fmt.Errorf("failed to store new incident: %w", err)
	} else if !saved {
		return nil, errors.New("failed to store new incident")
	}

	// Update Headers
	if err := s.updateHeader(incdnt); err != nil {
		return nil, fmt.Errorf("failed to update headers: %w", err)
	}

	return incdnt, nil
}

// UpdateIncident updates an incident.
func (s *incidentStore) UpdateIncident(incdnt *incident.Incident) error {
	if incdnt == nil {
		return errors.New("incident is nil")
	}
	if incdnt.ID == "" {
		return errors.New("ID should be set")
	}

	headers, err := s.getIDHeaders()
	if err != nil {
		return fmt.Errorf("failed to get all headers value: %w", err)
	}

	if _, exists := headers[incdnt.ID]; !exists {
		return fmt.Errorf("incident with id (%s) does not exist", incdnt.ID)
	}

	saved, err := s.pluginAPI.KV.Set(toIncidentKey(incdnt.ID), incdnt)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	} else if !saved {
		return errors.New("failed to update incident")
	}

	// Update Headers
	if err := s.updateHeader(incdnt); err != nil {
		return fmt.Errorf("failed to update headers: %w", err)
	}

	return nil
}

// GetIncident Gets an incident by ID.
func (s *incidentStore) GetIncident(id string) (*incident.Incident, error) {
	headers, err := s.getIDHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get all headers value: %w", err)
	}

	if _, exists := headers[id]; !exists {
		return nil, fmt.Errorf("incident with id (%s) does not exist: %w", id, incident.ErrNotFound)
	}

	return s.getIncident(id)
}

// GetIncidentIDForChannel Gets an incident associated to the given channel id.
func (s *incidentStore) GetIncidentIDForChannel(channelID string) (string, error) {
	headers, err := s.getIDHeaders()
	if err != nil {
		return "", fmt.Errorf("failed to get all headers value: %w", err)
	}

	// Search for which incident has the given channel associated
	for _, header := range headers {
		incdnt, err := s.getIncident(header.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get incident for id (%s): %w", header.ID, err)
		}

		for _, incidentChannelID := range incdnt.ChannelIDs {
			if incidentChannelID == channelID {
				return incdnt.ID, nil
			}
		}
	}
	return "", fmt.Errorf("channel with id (%s) does not have an incident: %w", channelID, incident.ErrNotFound)
}

// NukeDB Removes all incident related data.
func (s *incidentStore) NukeDB() error {
	return s.pluginAPI.KV.DeleteAll()
}

// toIncidentKey converts an incident to an internal key used to store in the KV Store.
func toIncidentKey(incidentID string) string {
	return incidentKey + incidentID
}

func toHeader(headers idHeaderMap) []incident.Header {
	var result []incident.Header
	for _, value := range headers {
		result = append(result, value)
	}

	return result
}

func (s *incidentStore) getIncident(incidentID string) (*incident.Incident, error) {
	var incdnt incident.Incident
	if err := s.pluginAPI.KV.Get(toIncidentKey(incidentID), &incdnt); err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}
	if incdnt.ID == "" {
		return nil, incident.ErrNotFound
	}
	return &incdnt, nil
}

func (s *incidentStore) getIDHeaders() (idHeaderMap, error) {
	headers := idHeaderMap{}
	if err := s.pluginAPI.KV.Get(allHeadersKey, &headers); err != nil {
		return nil, fmt.Errorf("failed to get all headers value: %w", err)
	}
	return headers, nil
}

func (s *incidentStore) updateHeader(incdnt *incident.Incident) error {
	headers, err := s.getIDHeaders()
	if err != nil {
		return fmt.Errorf("failed to get all headers: %w", err)
	}

	headers[incdnt.ID] = incdnt.Header

	// TODO: Should be using CompareAndSet, but deep copy is expensive.
	if saved, err := s.pluginAPI.KV.Set(allHeadersKey, headers); err != nil {
		return fmt.Errorf("failed to set all headers value: %w", err)
	} else if !saved {
		return errors.New("failed to set all headers value")
	}

	return nil
}

func headerMatchesFilter(header incident.Header, options incident.HeaderFilterOptions) bool {
	if options.TeamID != "" {
		return header.TeamID == options.TeamID
	}

	return true
}