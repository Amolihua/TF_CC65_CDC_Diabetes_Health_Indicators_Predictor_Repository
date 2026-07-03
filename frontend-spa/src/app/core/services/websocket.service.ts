import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  private socket!: WebSocket;
  public metrics$ = new Subject<any>();

  constructor() {
    this.connect();
  }

  private connect() {
    this.socket = new WebSocket('ws://localhost:8080/api/ws/metrics');
    
    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.metrics$.next(data);
      } catch (e) {
        console.error('[WS] Error de parseo', e);
      }
    };

    this.socket.onclose = () => {
      // Re-conexión resiliente
      setTimeout(() => this.connect(), 3000);
    };
  }
}
