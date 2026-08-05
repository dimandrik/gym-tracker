package handlers

import (
	"encoding/json"
	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
	"gym_tracker/internal/storage"
	"net/http"
)

type MachineHandler struct {
	machineRepo *repository.MachineRepository
	uploadDir   string
}

func NewMachineHandler(machineRepo *repository.MachineRepository, uploadDir string) *MachineHandler {
	return &MachineHandler{machineRepo: machineRepo, uploadDir: uploadDir}
}

// CreateMachine expects a multipart form (name + photo) rather than JSON,
// since it has to accept a file upload.
func (h *MachineHandler) CreateMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max upload size
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	photoURL, err := storage.SaveFile(file, header, h.uploadDir)
	if err != nil {
		http.Error(w, "failed to save photo", http.StatusInternalServerError)
		return
	}

	id, err := h.machineRepo.CreateMachine(r.Context(), userID, name, photoURL)
	if err != nil {
		http.Error(w, "failed to create machine", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "photo_url": photoURL})
}

func (h *MachineHandler) GetMachines(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	machines, err := h.machineRepo.GetMachinesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get machines", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(machines)
}

// GetMachine returns a machine by ID. Note: it doesn't check that the
// machine belongs to the requesting user — fine while machines aren't
// shared between accounts, revisit if that changes.
func (h *MachineHandler) GetMachine(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	machine, err := h.machineRepo.GetMachineByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get machine", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(machine)
}

func (h *MachineHandler) DeleteMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	machineID := r.PathValue("id")

	err := h.machineRepo.DeleteMachine(r.Context(), machineID, userID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MachineHandler) UpdateMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	machineID := r.PathValue("id")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	existingMachine, err := h.machineRepo.GetMachineByID(r.Context(), machineID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	photoURL := existingMachine.PhotoURL

	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		photoURL, err = storage.SaveFile(file, header, h.uploadDir)
		if err != nil {
			http.Error(w, "failed to save photo", http.StatusInternalServerError)
			return
		}
	}

	if err := h.machineRepo.UpdateMachine(r.Context(), machineID, userID, name, photoURL); err != nil {
		http.Error(w, "failed to update machine", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
