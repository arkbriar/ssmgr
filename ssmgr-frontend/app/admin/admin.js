'use strict';

angular.module('ssmgr.admin', ['ngRoute', 'ngMaterial'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/admin', {
    templateUrl: 'admin/admin.html',
    controller: 'AdminCtrl'
  });
}])

.controller('AdminCtrl', [function() {

}]);