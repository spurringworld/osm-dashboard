
import {HttpClient} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {ICanDeactivate} from '@common/interfaces/candeactivate';
import {NamespaceService} from '@common/services/global/namespace';
import {validateUniqueName} from '@create/from/form/validator/uniquename.validator';
import {FormValidators} from '@create/from/form/validator/validators';
import {CreateNamespaceDialog} from '@create/from/form/createnamespace/dialog';
import {DeployLabel} from '@create/from/form/deploylabel/deploylabel';
import {take, takeUntil} from 'rxjs/operators';
import {ActivatedRoute, Router} from '@angular/router';
import {MatDialog} from '@angular/material/dialog';
import {
  AppDeploymentSpec,
  EnvironmentVariable,
  Namespace,
  NamespaceList,
  PortMapping,
  Protocols,
  SecretList,
} from '@api/root.api';
import {Subject} from 'rxjs';

// Label keys for predefined labels
const APP_LABEL_KEY = 'k8s-app';

@Component({
  selector: 'kd-dialog-form',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class DialogFormComponent extends ICanDeactivate implements OnInit, OnDestroy {
  namespaces: string[];
  form: FormGroup;
  readonly nameMaxLength = 24;
  labelArr: DeployLabel[] = [];
  private created_ = false;
  private unsubscribe_ = new Subject<void>();

  constructor(
    private readonly namespace_: NamespaceService,
    private readonly http_: HttpClient,
    private readonly fb_: FormBuilder,
    private readonly dialog_: MatDialog,
    private readonly route_: ActivatedRoute,
  ) {
    super();
  }

  get name(): AbstractControl {
    return this.form.get('name');
  }

  get namespace(): AbstractControl {
    return this.form.get('namespace');
  }

  get labels(): FormArray {
    return this.form.get('labels') as FormArray;
  }

  ngOnInit(): void {
    this.form = this.fb_.group({
      name: ['', Validators.compose([Validators.required, FormValidators.namePattern])],
      namespace: [this.route_.snapshot.params.namespace || '', Validators.required],
    });
    this.labelArr = [new DeployLabel(APP_LABEL_KEY, '', false), new DeployLabel()];
    this.name.valueChanges.subscribe(v => {
      this.labelArr[0].value = v;
      this.labels.patchValue([{index: 0, value: v}]);
    });
    this.namespace.valueChanges.pipe(takeUntil(this.unsubscribe_)).subscribe((namespace: string) => {
      this.name.clearAsyncValidators();
      this.name.setAsyncValidators(validateUniqueName(this.http_, namespace));
      this.name.updateValueAndValidity();
    });
    this.http_.get('api/v1/namespace').subscribe((result: NamespaceList) => {
      this.namespaces = result.namespaces.map((namespace: Namespace) => namespace.objectMeta.name);
      this.namespace.patchValue(
        !this.namespace_.areMultipleNamespacesSelected()
          ? this.route_.snapshot.params.namespace || this.namespaces[0]
          : this.namespaces[0]
      );
      this.form.markAsPristine();
    });
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
  }

  hasUnsavedChanges(): boolean {
    return this.form.dirty;
  }

  isCreateDisabled(): boolean {
    return !this.form.valid;
  }

  areMultipleNamespacesSelected(): boolean {
    return this.namespace_.areMultipleNamespacesSelected();
  }

  handleNamespaceDialog(): void {
    const dialogData = {data: {namespaces: this.namespaces}};
    const dialogDef = this.dialog_.open(CreateNamespaceDialog, dialogData);
    dialogDef
      .afterClosed()
      .pipe(take(1))
      .subscribe(answer => {
        /**
         * Handles namespace dialog result. If namespace was created successfully then it
         * will be selected, otherwise first namespace will be selected.
         */
        if (answer) {
          this.namespaces.push(answer);
          this.namespace.patchValue(answer);
        } else {
          this.namespace.patchValue(this.namespaces[0]);
        }
      });
  }

  isVariableFilled(variable: EnvironmentVariable): boolean {
    return !!variable.name;
  }

  isNumber(value: string): boolean {
    return typeof value === 'number' && !isNaN(value);
  }

  deploy(): void {
    const spec = this.getSpec();
	  this.created_ = true;
  }

  private getSpec() {
    return {
      name: this.name.value,
      namespace: this.namespace.value,
    };
  }

  canDeactivate(): boolean {
    return this.form.pristine || this.created_;
  }
}
