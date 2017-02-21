'use strict';

angular.module('ssmgr.view.manager', ['ngMaterial', 'ngRoute'])

.config(['$routeProvider', ($routeProvider) => {
  $routeProvider.when('/manager', {
    templateUrl: 'views/manager/manager.html',
    controller: 'managerCtrl'
  });
}])

.controller('managerCtrl', [() => {

}])

.controller('navCtrl', ($scope, $mdSidenav, $log) => {
  $scope.close = () => {
    $mdSidenav('manager-nav').close()
    .then(() => {
      $log.debug("close manager nav is done");
    })
  };
})