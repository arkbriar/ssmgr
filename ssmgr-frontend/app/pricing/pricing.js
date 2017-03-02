'use strict';

angular.module('ssmgr.pricing', ['ngRoute', 'ngMaterial'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/pricing', {
    templateUrl: 'pricing/pricing.html',
    controller: 'PricingCtrl'
  });
}])

.controller('PricingCtrl', ['$scope', function($scope) {
  $scope.products = [
    {
      price: '$10',
      details: "$10 product",
      height: 300,
    },
    {
      price: '$20',
      details: "$20 product",
      height: 400,
    },
    {
      price: '$30',
      details: "$30 product",
      height: 350,
    },
  ];

}]);