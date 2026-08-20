package handlers

import (
	"encoding/json"
	"errors"
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

// принимает multipart-форму, а не JSON, т.к. нужно загрузить файл
func (h *MachineHandler) CreateMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	if err := r.ParseMultipartForm(10 << 20); err != nil { // лимит 10 МБ
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
		if errors.Is(err, storage.ErrInvalidFileType) {
			http.Error(w, "invalid file type, only JPEG, PNG, GIF, and WEBP images are allowed", http.StatusBadRequest)
			return
		}
		serverError(w, err, "failed to save photo")
		return
	}

	id, err := h.machineRepo.CreateMachine(r.Context(), userID, name, photoURL)
	if err != nil {
		serverError(w, err, "failed to create machine")
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
		serverError(w, err, "failed to get machines")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(machines)
}

func (h *MachineHandler) GetMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	machine, err := h.machineRepo.GetMachineByID(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(machine)
}

func (h *MachineHandler) DeleteMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	machineID := r.PathValue("id")

	machine, err := h.machineRepo.GetMachineByID(r.Context(), machineID, userID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	err = h.machineRepo.DeleteMachine(r.Context(), machineID, userID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	storage.DeleteFile(machine.PhotoURL, h.uploadDir)
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

	existingMachine, err := h.machineRepo.GetMachineByID(r.Context(), machineID, userID)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	photoURL := existingMachine.PhotoURL

	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		newPhotoURL, err := storage.SaveFile(file, header, h.uploadDir)
		if err != nil {
			if errors.Is(err, storage.ErrInvalidFileType) {
				http.Error(w, "invalid file type, only JPEG, PNG, GIF, and WEBP images are allowed", http.StatusBadRequest)
				return
			}
			serverError(w, err, "failed to save photo")
			return
		}
		oldPhotoURL := photoURL
		photoURL = newPhotoURL
		defer storage.DeleteFile(oldPhotoURL, h.uploadDir)
	}

	if err := h.machineRepo.UpdateMachine(r.Context(), machineID, userID, name, photoURL); err != nil {
		serverError(w, err, "failed to update machine")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
