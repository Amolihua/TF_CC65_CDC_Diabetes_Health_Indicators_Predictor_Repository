import { Component, OnInit, OnDestroy, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Subscription } from "rxjs";
import { WebsocketService } from "../../services/websocket.service";
import { ApiService } from "../../core/services/api.service";

@Component({
  selector: "app-dashboard",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./dashboard.component.html",
})
export class DashboardComponent implements OnInit, OnDestroy {
  private wsService = inject(WebsocketService);
  private apiService = inject(ApiService);
  private wsSubscription!: Subscription;

  //Variables operativas de estado del clúster
  stats = {
    cache_hits: 0,
    cache_misses: 0,
    cache_errors: 0,
    modelo_version: 0,
    nodos_activos: 0,
    cpu_goroutines: 0,
    ram_sys_mb: 0
  };
  selectedFile: File | null = null;
  isTraining = false;
  trainingMessage: string | null = null;

  ngOnInit() {
    this.wsSubscription = this.wsService.metrics$.subscribe((data: any) => {
      this.stats = data;
    });
  }

  onFileSelected(event: any) {
    this.selectedFile = event.target.files[0];
  }

  dispararEntrenamiento() {
    this.isTraining = true;
    this.trainingMessage = "Cargando archivo y ejecutando pipeline concurrente en los nodos...";

    const formData = new FormData();
    formData.append("dataset", this.selectedFile!);

    this.apiService.train(formData).subscribe({
      next: (res: any) => {
        this.trainingMessage = "¡Entrenamiento completado exitosamente! El clúster se actualizó a la nueva versión.";
        this.isTraining = false;
        this.selectedFile = null;
      },
      error: () => {
        this.trainingMessage = "Error en la orquestación del dataset. Verifique los logs del clúster Go.";
        this.isTraining = false;
      },
    });
  }

  ngOnDestroy() {
    if (this.wsSubscription) {
      this.wsSubscription.unsubscribe();
    }
  }
}
