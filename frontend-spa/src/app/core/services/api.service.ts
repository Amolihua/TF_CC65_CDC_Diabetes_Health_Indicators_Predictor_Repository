import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  // Endpoint para el admin
  login(credentials: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/login`, credentials);
  }

  // Endpoint de inferencia distribuida para el ciudadano
  predict(patientProfile: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/predict`, patientProfile);
  }

  // Endpoint de reentrenamiento para el admin
  train(dataset: FormData): Observable<any> {
    return this.http.post(`${this.apiUrl}/train`, dataset);
  }

  // Endpoint para metricas
  getMetrics(): Observable<any> {
    return this.http.get(`${this.apiUrl}/metrics`);
  }

  // Endpoint para historial
  getHistorial(): Observable<any> {
    return this.http.get(`${this.apiUrl}/historial`);
  }
}
