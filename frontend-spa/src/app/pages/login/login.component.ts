import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html'
})
export class LoginComponent {
  username = '';
  password = '';
  errorMessage = '';
  isLoading = false;

  private apiService = inject(ApiService);
  private router = inject(Router);

  login() {
    this.isLoading = true;
    this.errorMessage = '';
    
    this.apiService.login({
      username: this.username,
      password: this.password
    }).subscribe({
      next: (res) => {
        if (res && res.token) {
          localStorage.setItem('token', res.token);
          this.router.navigate(['/admin/dashboard']);
        } else {
          this.isLoading = false;
          this.errorMessage = 'El servidor no devolvió un token válido.';
        }
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = 'Credenciales inválidas o error de conexión.';
        console.error("Login error:", err);
      }
    });
  }
}
