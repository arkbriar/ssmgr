'use strict';

angular.module('ssmgr.view.user', ['ngMaterial'])

.config(['$routeProvider', ($routeProvider) => {
  $routeProvider.when('/user', {
    templateUrl: 'user/user.html',
    controller: 'userCtrl'
  });
}])

.controller('userCtrl', [() => {

}]);