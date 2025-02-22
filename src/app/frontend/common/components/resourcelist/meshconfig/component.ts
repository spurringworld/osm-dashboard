// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {Observable} from 'rxjs';
import {Meshconfig, MeshconfigList} from 'typings/root.api';

import {ResourceListWithStatuses} from '@common/resources/list';
import {NotificationsService} from '@common/services/global/notifications';
import {EndpointManager, Resource} from '@common/services/resource/endpoint';
import {NamespacedResourceService} from '@common/services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {Status, StatusClass} from '../statuses';
import {VerberService} from '@common/services/global/verber';

@Component({
  selector: 'kd-meshconfig-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class MeshConfigListComponent extends ResourceListWithStatuses<MeshconfigList, Meshconfig> {
  @Input() endpoint = EndpointManager.resource(Resource.meshconfig, true).list();

  constructor(
    private readonly service_: NamespacedResourceService<MeshconfigList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
    private readonly verber_: VerberService
  ) {
    super('meshconfig', notifications, cdr);
    this.id = ListIdentifier.meshConfig;
    this.groupId = ListGroupIdentifier.osm;

    // Register status icon handlers
    this.registerBinding(StatusClass.Success, r => this.isInSuccessState(r), Status.Success);
    this.registerBinding(StatusClass.Warning, r => !this.isInSuccessState(r), Status.Pending);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<MeshconfigList> {
    return this.service_.get(this.endpoint, undefined, undefined, params);
  }

  map(meshconfigList: MeshconfigList): Meshconfig[] {
    return meshconfigList.meshconfigs;
  }

  /**
   * Success state of a Service depends on the type of service
   * https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
   * ClusterIP:     ClusterIP is defined
   * NodePort:      ClusterIP is defined
   * LoadBalancer:  ClusterIP is defined __and__ external endpoints exist
   * ExternalName:  true
   */
  isInSuccessState(resource: Meshconfig): boolean {
    return !!resource;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'created'];
  }

  onInstall(): void {
    this.verber_.showInstallDialog('',null,null);
  }
  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
