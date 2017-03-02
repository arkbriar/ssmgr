'use strict';

angular.module('ssmgr.user', ['ngRoute', 'ngMaterial'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/user', {
    templateUrl: 'user/user.html',
    controller: 'UserCtrl'
  });
}])

.controller('UserCtrl', [function() {

}]);