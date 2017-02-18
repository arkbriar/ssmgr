'use strict';

// Declare app level module which depends on views, and components
angular.module('myApp', [
  'ngRoute',

  'hello',
  'about',
  'user',
  'manager'
]).
config(['$locationProvider', '$routeProvider', function($locationProvider, $routeProvider) {
  $locationProvider.hashPrefix('!');

  $routeProvider.otherwise({redirectTo: 'index.html'});
}]);
