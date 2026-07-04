import { Component, OnInit, OnDestroy, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Subscription } from "rxjs";
import { WebsocketService } from "../../services/websocket.service";
import { ApiService } from "../../core/services/api.service";
import Chart from 'chart.js/auto';

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
  
  nodos: any[] = [];
  charts: { [hostname: string]: Chart } = {};
  historial: any[] = [];

  ngOnInit() {
    this.wsSubscription = this.wsService.metrics$.subscribe((data: any) => {
      this.stats = data;
      
      if (data.cluster_nodes) {
        this.nodos = data.cluster_nodes.sort((a: any, b: any) => a.hostname.localeCompare(b.hostname));
        setTimeout(() => this.updateCharts(), 0);
      }
    });

    this.cargarHistorial();
  }

  cargarHistorial() {
    this.apiService.getHistorial().subscribe({
      next: (data) => {
        this.historial = data;
      },
      error: (err) => console.error("Error al cargar historial", err)
    });
  }

  trackByHostname(index: number, nodo: any): string {
    return nodo.hostname;
  }

  updateCharts() {
    this.nodos.forEach(nodo => {
      const canvasId = `chart-${nodo.hostname}`;
      const canvas = document.getElementById(canvasId) as HTMLCanvasElement;
      
      if (!canvas) return;

      if (!this.charts[nodo.hostname]) {
        this.charts[nodo.hostname] = new Chart(canvas, {
          type: 'line',
          data: {
            labels: Array(50).fill(''),
            datasets: [
              {
                label: 'RAM Asignada (OS) [MB]',
                data: Array(50).fill(0),
                borderColor: '#10b981',
                backgroundColor: 'rgba(16, 185, 129, 0.1)',
                tension: 0.3,
                borderWidth: 2,
                pointRadius: 0,
                fill: true
              },
              {
                label: 'RAM Viva (Alloc) [MB]',
                data: Array(50).fill(0),
                borderColor: '#f59e0b',
                backgroundColor: 'rgba(245, 158, 11, 0.1)',
                tension: 0.3,
                borderWidth: 2,
                pointRadius: 0,
                fill: true
              },
              {
                label: 'CPU (Goroutines)',
                data: Array(50).fill(0),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                tension: 0.3,
                borderWidth: 2,
                pointRadius: 0,
                fill: true
              }
            ]
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: { mode: 'index', intersect: false },
            scales: {
              y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#94a3b8' } },
              x: { grid: { display: false } }
            },
            plugins: { legend: { labels: { color: '#e2e8f0' } } }
          }
        });
      }

      const chart = this.charts[nodo.hostname];
      const ramSysData = chart.data.datasets[0].data as number[];
      const ramAllocData = chart.data.datasets[1].data as number[];
      const cpuData = chart.data.datasets[2].data as number[];
      
      ramSysData.shift();
      ramSysData.push(nodo.ram_sys_mb);
      
      ramAllocData.shift();
      ramAllocData.push(nodo.ram_alloc_mb || 0);
      
      cpuData.shift();
      cpuData.push(nodo.goroutines);
      
      chart.update();
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
