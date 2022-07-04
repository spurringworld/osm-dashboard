
import {HttpClient} from '@angular/common/http';
import {Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
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

import {MatButtonToggleGroup} from '@angular/material/button-toggle';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';
import {dump as toYaml, load as fromYaml} from 'js-yaml';
import {EditorMode} from '../../../components/textinput/component';
import {meshOptions} from './meshOptions';

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
  selector: 'kd-install-resource-form',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class DialogFormComponent extends ICanDeactivate implements OnInit, OnDestroy {
  namespaces: string[];
	options: string[];
  form: FormGroup;
  readonly nameMaxLength = 24;
  labelArr: DeployLabel[] = [];
  private created_ = false;
  private unsubscribe_ = new Subject<void>();
  selectedMode = EditorMode.YAML;
  @ViewChild('group', {static: true}) buttonToggleGroup: MatButtonToggleGroup;
  modes = EditorMode;
  text = toYaml(meshOptions);

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
	
	get timeout(): AbstractControl {
	  return this.form.get('timeout');
	}
	
  get chartPath(): AbstractControl {
    return this.form.get('chartPath');
  }
	
	get atomic(): AbstractControl {
	  return this.form.get('atomic');
	}
	
	get enforceSingleMesh(): AbstractControl {
	  return this.form.get('enforceSingleMesh');
	}
	
	get tracingAddress(): AbstractControl {
	  return this.form.get('tracingAddress');
	}
	
	get tracingPort(): AbstractControl {
	  return this.form.get('tracingPort');
	}
	
	get tracingEndpoint(): AbstractControl {
	  return this.form.get('tracingEndpoint');
	}
	
	get tracingEnabled(): AbstractControl {
	  return this.form.get('tracingEnabled');
	}
	
	get tracingDeploy(): AbstractControl {
	  return this.form.get('tracingDeploy');
	}
	
	get metricsAddress(): AbstractControl {
	  return this.form.get('metricsAddress');
	}
	
  get metricsPort(): AbstractControl {
    return this.form.get('metricsPort');
  }
	
  get metricsDeploy(): AbstractControl {
    return this.form.get('metricsDeploy');
  }

  ngOnInit(): void {
    this.form = this.fb_.group({
      name: ['osm', Validators.compose([Validators.required, FormValidators.namePattern])],
      namespace: [this.route_.snapshot.params.namespace || '', Validators.required],
			enforceSingleMesh: [true],
			atomic: [false],
			tracingEnabled: [meshOptions.osm.tracing.enable],
			tracingDeploy: [true],
			tracingAddress: [meshOptions.osm.tracing.address],
			tracingPort: [meshOptions.osm.tracing.port],
			tracingEndpoint: [meshOptions.osm.tracing.endpoint],
			metricsDeploy: [meshOptions.osm.deployPrometheus],
			metricsPort: [meshOptions.osm.prometheus.port],
			metricsAddress: [meshOptions.osm.prometheus.image],
			timeout: ['300s'],
    });
    this.labelArr = [new DeployLabel(APP_LABEL_KEY, '', false), new DeployLabel()];
    // this.name.valueChanges.subscribe(v => {
    //   this.labelArr[0].value = v;
    //   this.labels.patchValue([{index: 0, value: v}]);
    // });
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

  private getSpec() {
    return {
      name: this.name.value,
      namespace: this.namespace.value,
    };
  }

  canDeactivate(): boolean {
    return this.form.pristine || this.created_;
  }

  getJSON(): string {
    if (this.selectedMode === EditorMode.YAML) {
      return this.toRawJSON(fromYaml(this.text));
    }

    return this.text;
  }

	install(): string {
		let _payload = {
      name: this.name.value,
      namespace: this.namespace.value,
      timeout: this.timeout.value,
      atomic: this.atomic.value,
      enforceSingleMesh: this.enforceSingleMesh.value,
			options: JSON.parse(this.getJSON())
		};
	  this.created_ = true;
		return this.toRawJSON(_payload);
	}

  getSelectedMode(): string {
    return this.buttonToggleGroup.value;
  }

	updateOptionsTab(): void {
		let _options = meshOptions;
		if (this.selectedMode === EditorMode.YAML) {
			_options = fromYaml(this.text);
		} else {
			_options = this.text;
		}
		_options.osm.tracing.enable = this.form.get('tracingEnabled').value;
		_options.osm.tracing.address = this.form.get('tracingAddress').value;
		_options.osm.tracing.port = this.form.get('tracingPort').value;
		_options.osm.tracing.endpoint = this.form.get('tracingEndpoint').value;
		_options.osm.deployPrometheus = this.form.get('metricsDeploy').value;
		_options.osm.prometheus.port = this.form.get('metricsPort').value;
		_options.osm.prometheus.image = this.form.get('metricsAddress').value;
		
    if (this.selectedMode === EditorMode.YAML) {
      this.text = toYaml(_options);
    } else {
      this.text = this.toRawJSON(_options);
    }
	}
	
	updateBasicTab(): void {
		let _options = meshOptions;
		if (this.selectedMode === EditorMode.YAML) {
			_options = fromYaml(this.text);
		} else {
			_options = this.text;
		}
		this.form.get('tracingEnabled').setValue(_options.osm.tracing.enable, {emitEvent: false});
		this.form.get('tracingAddress').setValue(_options.osm.tracing.address, {emitEvent: false});
		this.form.get('tracingPort').setValue(_options.osm.tracing.port, {emitEvent: false});
		this.form.get('tracingEndpoint').setValue(_options.osm.tracing.endpoint, {emitEvent: false});
		this.form.get('metricsDeploy').setValue(_options.osm.deployPrometheus, {emitEvent: false});
		this.form.get('metricsPort').setValue(_options.osm.prometheus.port, {emitEvent: false});
		this.form.get('metricsAddress').setValue(_options.osm.prometheus.image, {emitEvent: false});
	}

  private updateText(): void {
    if (this.selectedMode === EditorMode.YAML) {
      this.text = toYaml(JSON.parse(this.text));
    } else {
      this.text = this.toRawJSON(fromYaml(this.text));
    }
  }

  private toRawJSON(object: {}): string {
    return JSON.stringify(object, null, '\t');
  }
}
