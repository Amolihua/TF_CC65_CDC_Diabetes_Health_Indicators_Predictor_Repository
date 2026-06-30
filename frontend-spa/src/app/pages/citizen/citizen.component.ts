import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';

@Component({
  selector: 'app-citizen',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './citizen.component.html'
})
export class CitizenComponent {
  private fb = inject(FormBuilder);
  private apiService = inject(ApiService);

  form: FormGroup;
  predictionResult: number | null = null;
  isLoading = false;
  errorMessage: string | null = null;

  constructor() {
    // Estructuración del formulario con enlace directo a las propiedades del backend
    this.form = this.fb.group({
      high_bp: [0, Validators.required],
      high_chol: [0, Validators.required],
      chol_check: [1, Validators.required],
      bmi: [25, [Validators.required, Validators.min(10), Validators.max(60)]],
      smoker: [0, Validators.required],
      stroke: [0, Validators.required],
      heart_disease_or_attack: [0, Validators.required],
      phys_activity: [1, Validators.required],
      fruits: [1, Validators.required],
      veggies: [1, Validators.required],
      hvy_alcohol_consump: [0, Validators.required],
      any_healthcare: [1, Validators.required],
      no_doc_bc_cost: [0, Validators.required],
      gen_hlth: [3, [Validators.required, Validators.min(1), Validators.max(5)]],
      ment_hlth: [0, [Validators.required, Validators.min(0), Validators.max(30)]],
      phys_hlth: [0, [Validators.required, Validators.min(0), Validators.max(30)]],
      diff_walk: [0, Validators.required],
      sex: [0, Validators.required],
      age: [5, Validators.required],
      education: [4, Validators.required],
      income: [5, Validators.required]
    });
  }

  enviarEvaluacion() {
    this.isLoading = true;
    this.predictionResult = null;
    this.errorMessage = null;

    const raw = this.form.value;
    const payload = {
      high_bp: Number(raw.high_bp), high_chol: Number(raw.high_chol),
      chol_check: Number(raw.chol_check), bmi: Number(raw.bmi),
      smoker: Number(raw.smoker), stroke: Number(raw.stroke),
      heart_disease_or_attack: Number(raw.heart_disease_or_attack),
      phys_activity: Number(raw.phys_activity), fruits: Number(raw.fruits),
      veggies: Number(raw.veggies), hvy_alcohol_consump: Number(raw.hvy_alcohol_consump),
      any_healthcare: Number(raw.any_healthcare), no_doc_bc_cost: Number(raw.no_doc_bc_cost),
      gen_hlth: Number(raw.gen_hlth), ment_hlth: Number(raw.ment_hlth),
      phys_hlth: Number(raw.phys_hlth), diff_walk: Number(raw.diff_walk),
      sex: Number(raw.sex), age: Number(raw.age),
      education: Number(raw.education), income: Number(raw.income)
    };

    this.apiService.predict(payload).subscribe({
      next: (res: any) => {
        this.predictionResult = res.prediction;
        this.isLoading = false;
      },
      error: () => {
        this.errorMessage = 'No se pudo conectar con el clúster. Reintente en unos momentos.';
        this.isLoading = false;
      }
    });
  }

  limpiarFormulario() {
    this.form.reset({
      high_bp: 0, high_chol: 0, chol_check: 1, bmi: 25, smoker: 0, stroke: 0,
      heart_disease_or_attack: 0, phys_activity: 1, fruits: 1, veggies: 1,
      hvy_alcohol_consump: 0, any_healthcare: 1, no_doc_bc_cost: 0, gen_hlth: 3,
      ment_hlth: 0, phys_hlth: 0, diff_walk: 0, sex: 0, age: 5, education: 4, income: 5
    });
    this.predictionResult = null;
    this.errorMessage = null;
  }
}
