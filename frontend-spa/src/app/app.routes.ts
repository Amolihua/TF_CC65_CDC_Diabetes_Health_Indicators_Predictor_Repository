import { Routes } from '@angular/router';
import { CitizenComponent } from './pages/citizen/citizen.component';
import { LoginComponent } from './pages/login/login.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';

export const routes: Routes = [
  { path: "", component: CitizenComponent },
  { path: "login", component: LoginComponent },
  { path: "admin/dashboard", component: DashboardComponent },
  { path: "**", redirectTo: "" },
];
