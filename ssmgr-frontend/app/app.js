'use strict';

// Declare app level module which depends on views, and components
angular.module('ssmgr', [
  'ngRoute',
  'ssmgr.views'
]).
config(['$locationProvider', '$routeProvider', function($locationProvider, $routeProvider) {
  $locationProvider.hashPrefix('!');

  $routeProvider.otherwise({redirectTo: '/user'});
}]);
