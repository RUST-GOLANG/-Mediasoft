package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Car struct {
	ID         int     `json:"id,omitempty"`
	Brand      string  `json:"brand,omitempty"`
	Model      string  `json:"model,omitempty"`
	Mileage    float64 `json:"mileage,omitempty"`
	OwnerCount int     `json:"owner_count,omitempty"`
}

var (
	cars      []Car
	NextID    = 1
	CarsMutex sync.Mutex
	filename  = "cars.json"
)

func LoadCars() {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			cars = []Car{}
			return
		}
		log.Fatalf("Error loading cars: %v", err)
	}
	json.Unmarshal(file, &cars)
	if len(cars) > 0 {
		NextID = cars[len(cars)-1].ID + 1
	}
}

func saveCars() error {
	data, err := json.Marshal(cars)
	if err != nil {
		log.Fatalf("Error saving cars: %v", err)
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Error saving cars: %v", err)
		return err
	}
	return nil
}

func getCars(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id != "" {
		for _, car := range cars {
			if strconv.Itoa(car.ID) == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(car)
				return
			}
		}
		http.Error(w, "Car not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cars)
}

func createCar(w http.ResponseWriter, r *http.Request) {
	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		log.Printf("Ошибка декодирования машины: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	CarsMutex.Lock()
	defer CarsMutex.Unlock()
	car.ID = NextID
	NextID++
	cars = append(cars, car)
	if err := saveCars(); err != nil {
		log.Printf("Ошибка сохранения машин: %v", err)
		http.Error(w, "Ошибка сохранения машины", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func updateCar(w http.ResponseWriter, r *http.Request) {
	var updatedCar Car
	if err := json.NewDecoder(r.Body).Decode(&updatedCar); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	CarsMutex.Lock()
	defer CarsMutex.Unlock()
	for i, car := range cars {
		if car.ID == updatedCar.ID {
			if updatedCar.Brand != "" {
				car.Brand = updatedCar.Brand
			}
			if updatedCar.Model != "" {
				car.Model = updatedCar.Model
			}
			if updatedCar.Mileage != 0 {
				car.Mileage = updatedCar.Mileage
			}
			if updatedCar.OwnerCount != 0 {
				car.OwnerCount = updatedCar.OwnerCount
			}
			cars[i] = car
			saveCars()
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(car)
			return
		}
	}
	http.Error(w, "Car not found", http.StatusNotFound)
}

func deleteCar(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	CarsMutex.Lock()
	defer CarsMutex.Unlock()
	for i, car := range cars {
		if strconv.Itoa(car.ID) == id {
			cars = append(cars[:i], cars[i+1:]...)
			saveCars()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Car not found", http.StatusNotFound)
}

func main() {
	LoadCars()
	http.HandleFunc("/cars", getCars)
	http.HandleFunc("/car/create", createCar)
	http.HandleFunc("/car/update", updateCar)
	http.HandleFunc("/car/delete", deleteCar)
	http.Handle("/", http.FileServer(http.Dir("./"))) 
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
